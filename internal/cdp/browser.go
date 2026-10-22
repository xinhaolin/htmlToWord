// Package cdp provides the small Chrome DevTools transport used by htmlword.
// It talks to a dedicated headless Chrome over OS pipes, never a user's browser.
package cdp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
)

type Message struct {
	ID        int64           `json:"id,omitempty"`
	Method    string          `json:"method,omitempty"`
	Params    json.RawMessage `json:"params,omitempty"`
	SessionID string          `json:"sessionId,omitempty"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type Browser struct {
	cmd       *exec.Cmd
	input     *os.File
	output    *os.File
	dir       string
	mu        sync.Mutex
	writeMu   sync.Mutex
	pending   map[int64]chan Message
	next      atomic.Int64
	done      chan struct{}
	onEvent   func(Message)
	closeOnce sync.Once
	stderr    bytes.Buffer
}

func FindChrome() (string, error) {
	return FindChromePath(os.Getenv("HTMLWORD_CHROME"))
}

// FindChromePath resolves an explicit executable without changing process state.
func FindChromePath(explicit string) (string, error) {
	if explicit != "" {
		p, err := exec.LookPath(explicit)
		if err != nil {
			return "", fmt.Errorf("Chrome executable: %w", err)
		}
		return p, nil
	}
	candidates := []string{"google-chrome", "chromium", "chromium-browser", "chrome"}
	if runtime.GOOS == "darwin" {
		candidates = append([]string{"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", "/Applications/Chromium.app/Contents/MacOS/Chromium"}, candidates...)
	}
	if runtime.GOOS == "windows" {
		for _, base := range []string{os.Getenv("PROGRAMFILES"), os.Getenv("PROGRAMFILES(X86)"), os.Getenv("LOCALAPPDATA")} {
			candidates = append(candidates, filepath.Join(base, "Google", "Chrome", "Application", "chrome.exe"))
		}
	}
	for _, p := range candidates {
		if v, e := exec.LookPath(p); e == nil {
			return v, nil
		}
	}
	return "", errors.New("Chrome/Chromium was not found; install it or pass an explicit executable path")
}

func Start(ctx context.Context, onEvent func(Message)) (*Browser, error) {
	return StartWithOptions(ctx, Options{Path: os.Getenv("HTMLWORD_CHROME"), NoSandbox: os.Getenv("HTMLWORD_NO_SANDBOX") == "1"}, onEvent)
}

type Options struct {
	Path, TempDir string
	NoSandbox     bool
}

func StartWithOptions(ctx context.Context, options Options, onEvent func(Message)) (*Browser, error) {
	if runtime.GOOS == "windows" {
		return nil, errors.New("Windows CDP pipe transport is not supported")
	}
	p, e := FindChromePath(options.Path)
	if e != nil {
		return nil, e
	}
	dir, e := os.MkdirTemp(options.TempDir, "htmlword-chrome-*")
	if e != nil {
		return nil, e
	}
	toChromeR, toChromeW, e := os.Pipe()
	if e != nil {
		os.RemoveAll(dir)
		return nil, e
	}
	fromChromeR, fromChromeW, e := os.Pipe()
	if e != nil {
		toChromeR.Close()
		toChromeW.Close()
		os.RemoveAll(dir)
		return nil, e
	}
	args := []string{"--headless=new", "--remote-debugging-pipe", "--user-data-dir=" + dir, "--no-first-run", "--no-default-browser-check", "--disable-background-networking", "--disable-sync", "--disable-extensions", "--disable-component-update", "--disable-default-apps", "--hide-scrollbars", "--mute-audio", "--disable-dev-shm-usage", "about:blank"}
	if options.NoSandbox {
		args = append(args, "--no-sandbox")
	}
	b := &Browser{dir: dir, input: toChromeW, output: fromChromeR, pending: make(map[int64]chan Message), done: make(chan struct{}), onEvent: onEvent}
	b.cmd = exec.CommandContext(ctx, p, args...)
	b.cmd.ExtraFiles = []*os.File{toChromeR, fromChromeW}
	b.cmd.Stderr = &b.stderr
	if e = b.cmd.Start(); e != nil {
		toChromeR.Close()
		toChromeW.Close()
		fromChromeR.Close()
		fromChromeW.Close()
		os.RemoveAll(dir)
		return nil, e
	}
	toChromeR.Close()
	fromChromeW.Close()
	go b.readLoop()
	return b, nil
}

func (b *Browser) readLoop() {
	defer close(b.done)
	r := bufio.NewReaderSize(b.output, 64*1024)
	for {
		line, e := r.ReadBytes(0)
		if e != nil {
			return
		}
		var m Message
		if json.Unmarshal(bytes.TrimSuffix(line, []byte{0}), &m) != nil {
			continue
		}
		if m.ID != 0 {
			b.mu.Lock()
			ch := b.pending[m.ID]
			delete(b.pending, m.ID)
			b.mu.Unlock()
			if ch != nil {
				ch <- m
			}
		} else if b.onEvent != nil {
			b.onEvent(m)
		}
	}
}

func (b *Browser) Call(ctx context.Context, session, method string, params any, result any) error {
	id := b.next.Add(1)
	ch := make(chan Message, 1)
	p, e := json.Marshal(params)
	if e != nil {
		return e
	}
	data, e := json.Marshal(Message{ID: id, Method: method, Params: p, SessionID: session})
	if e != nil {
		return e
	}
	b.mu.Lock()
	b.pending[id] = ch
	b.mu.Unlock()
	defer func() { b.mu.Lock(); delete(b.pending, id); b.mu.Unlock() }()
	b.writeMu.Lock()
	_, e = b.input.Write(append(data, 0))
	b.writeMu.Unlock()
	if e != nil {
		return e
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-b.done:
		b.Close()
		return fmt.Errorf("Chrome 提前退出：%s", b.stderr.String())
	case m := <-ch:
		if m.Error != nil {
			return fmt.Errorf("%s: %s", method, m.Error.Message)
		}
		if result != nil {
			return json.Unmarshal(m.Result, result)
		}
		return nil
	}
}

func (b *Browser) Eval(ctx context.Context, session, expression string, result any) error {
	var r struct {
		Result struct {
			Value       json.RawMessage `json:"value"`
			Description string          `json:"description"`
		} `json:"result"`
		Exception json.RawMessage `json:"exceptionDetails"`
	}
	if e := b.Call(ctx, session, "Runtime.evaluate", map[string]any{"expression": expression, "returnByValue": true, "awaitPromise": true}, &r); e != nil {
		return e
	}
	if len(r.Exception) > 0 {
		return fmt.Errorf("页面分析失败: %s %s", r.Result.Description, string(r.Exception))
	}
	if result != nil {
		if len(r.Result.Value) == 0 {
			return errors.New("浏览器未返回分析结果")
		}
		return json.Unmarshal(r.Result.Value, result)
	}
	return nil
}

func (b *Browser) Close() {
	b.closeOnce.Do(func() {
		b.input.Close()
		if b.cmd.Process != nil {
			b.cmd.Process.Kill()
		}
		b.cmd.Wait()
		b.output.Close()
		os.RemoveAll(b.dir)
	})
}
