// Materialize the same sample HTML used by the browser, without a build step.
import fs from 'node:fs';
import path from 'node:path';
import vm from 'node:vm';
import {fileURLToPath} from 'node:url';
const root=path.resolve(path.dirname(fileURLToPath(import.meta.url)),'..');
const source=fs.readFileSync(path.join(root,'tools/playground/internal/server/web/app.js'),'utf8');
const complex=fs.readFileSync(path.join(root,'tools/playground/internal/server/web/complex-samples.js'),'utf8');
const resume=fs.readFileSync(path.join(root,'tools/playground/internal/server/web/resume-sample.js'),'utf8');
const editorial=fs.readFileSync(path.join(root,'tools/playground/internal/server/web/resume-editorial-sample.js'),'utf8');
const batch=fs.readFileSync(path.join(root,'tools/playground/internal/server/web/batch-samples.js'),'utf8');
const samples=vm.runInNewContext(complex+'\n'+resume+'\n'+editorial+'\n'+batch+'\n'+source.slice(0,source.indexOf("let editor='html'"))+';samples',{btoa,encodeURIComponent});
for(const [name,html]of Object.entries(samples))fs.writeFileSync(path.join(root,'testdata/html',name+'.html'),html);
console.log('Exported '+Object.keys(samples).length+' sample HTML files');
