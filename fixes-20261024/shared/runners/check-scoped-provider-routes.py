from pathlib import Path
import os,sys,json,subprocess,tempfile,time,signal,selectors
b=Path.home()/'.ao/dev/harness-fixes-6541-6546-20261023'
print('READY_FOR_SCOPED_INPUT',flush=True)
keys=json.loads(sys.stdin.readline())
assert set(keys)=={'zai','minimax'} and all(isinstance(v,str) and v for v in keys.values())
node=str(Path((b/'node24-path.txt').read_text().strip())/'node')
base=b/'ephemeral-provider-checks'
base.mkdir(exist_ok=True)
native=b/'native/node_modules/.bin'
probe=(b/'mastra/backend/internal/adapters/agent/mastracode/auth-probe.mjs').read_text()
def run(argv,env,cwd,limit=25):
 p=subprocess.Popen(argv,env=env,cwd=cwd,stdin=subprocess.DEVNULL,stdout=subprocess.PIPE,stderr=subprocess.DEVNULL,start_new_session=True)
 sel=selectors.DefaultSelector();sel.register(p.stdout,selectors.EVENT_READ)
 raw=bytearray();start=time.monotonic();reason='completed'
 while True:
  if time.monotonic()-start>limit:reason='timeout';break
  events=sel.select(.1)
  if events:
   part=os.read(p.stdout.fileno(),8192)
   if not part:break
   raw.extend(part)
   if len(raw)>65536:reason='output_limit';break
  elif p.poll() is not None:break
 if reason!='completed':
  try:os.killpg(p.pid,signal.SIGTERM)
  except ProcessLookupError:pass
 try:rc=p.wait(timeout=3)
 except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);rc=p.wait()
 sel.close()
 return rc,bytes(raw),reason,round(time.monotonic()-start,2)
results=[]
for provider,model in [('zai','zai/glm-5.2'),('minimax','minimax/MiniMax-M3')]:
 with tempfile.TemporaryDirectory(prefix='mastra-'+provider+'-',dir=base) as td:
  d=Path(td);profile=d/'profile';home=d/'home';home.mkdir();tmp=d/'tmp';tmp.mkdir()
  env={'HOME':str(home),'PATH':str(Path(node).parent)+':/usr/bin:/bin','TMPDIR':str(tmp),'MASTRA_APP_DATA_DIR':str(profile),'ZHIPU_API_KEY' if provider=='zai' else 'MINIMAX_API_KEY':keys[provider]}
  setup="const m=await import(process.argv[1]);const s=m.loadSettings();s.models.activeModelPackId=null;s.models.modeDefaults={...s.models.modeDefaults,build:process.argv[2]};m.saveSettings(s);"
  rc,raw,reason,sec=run([node,'--input-type=module','-e',setup,'--',(b/'native/node_modules/@mastra/code-sdk/dist/onboarding/settings.js').as_uri(),model],env,d,10)
  r={'harness':'mastra-code','provider':provider,'model':model,'route':'native general API provider, not coding-plan tool emulation','setupExit':rc,'status':'INCONCLUSIVE'}
  if rc==0:
   rc,raw,reason,sec=run([node,'--input-type=module','-e',probe,'--',str(native/'mastracode'),str(profile)],env,d,25)
   r.update({'exit':rc,'termination':reason,'seconds':sec})
   try:
    val=json.loads(raw)
    if isinstance(val,dict) and set(val)<= {'status','code','provider','httpStatus'}:r['nativeProbe']=val;r['status']='PASS' if val.get('status')=='authorized' else 'UNVERIFIED'
   except:pass
  results.append(r);print(json.dumps(r),flush=True)
for provider,model,url in [('zai','glm-5.2','https://api.z.ai/api/paas/v4'),('minimax','MiniMax-M3','https://api.minimax.io/v1')]:
 with tempfile.TemporaryDirectory(prefix='codebuddy-'+provider+'-',dir=base) as td:
  d=Path(td);home=d/'home';home.mkdir();profile=d/'profile';profile.mkdir();tmp=d/'tmp';tmp.mkdir()
  env={'HOME':str(home),'PATH':str(Path(node).parent)+':/usr/bin:/bin','TMPDIR':str(tmp),'CODEBUDDY_CONFIG_DIR':str(profile),'CODEBUDDY_API_KEY':keys[provider],'CODEBUDDY_BASE_URL':url,'CODEBUDDY_MODEL':model,'CODEBUDDY_CODE_DISABLE_BACKGROUND_TASKS':'1','CODEBUDDY_DISABLE_CRON':'1','DISABLE_AUTOUPDATER':'1','CBC_ENABLE_FUNCTION_HOOKS':'0','CODEBUDDY_PLUGIN_DIRS':''}
  argv=[str(native/'codebuddy'),'--print','--output-format','json','--tools','','--max-turns','1','--no-session-persistence','--strict-mcp-config','--mcp-config','{"mcpServers":{}}','--setting-sources','user','--settings','{"disableAllHooks":true}','--agent','cli','--system-prompt','Reply briefly without using tools.','--','Reply with OK.']
  rc,raw,reason,sec=run(argv,env,d,32)
  r={'harness':'codebuddy','provider':provider,'model':model,'baseUrl':url,'exit':rc,'termination':reason,'seconds':sec,'status':'INCONCLUSIVE','structuredStatusCodes':[]}
  try:
   val=json.loads(raw)
   if isinstance(val,list):val=val[-1]
   if isinstance(val,dict) and val.get('type')=='result':
    ok=rc==0 and val.get('is_error') is False and val.get('subtype')=='success' and not val.get('terminal_reason') and not val.get('errors') and not val.get('errors_info') and isinstance(val.get('result'),str) and bool(val['result'].strip()) and val.get('usage',{}).get('output_tokens',0)>0
    r['status']='PASS' if ok else 'UNVERIFIED'
    r['structuredStatusCodes']=[x['status'] for x in val.get('errors_info',[]) if isinstance(x.get('status'),int)]
  except:pass
  results.append(r);print(json.dumps(r),flush=True)
# These files contain verdicts only. All native profiles/HOME directories are deleted.
out={'nativePackages':json.loads((b/'native-versions.json').read_text()),'sourceHead':subprocess.check_output(['git','-C',str(b/'mastra'),'rev-parse','HEAD'],text=True).strip(),'credentialsPersistedByRunner':False,'profilesRetained':False,'checks':results}
(b/'scoped-provider-route-results.json').write_text(json.dumps(out,indent=2)+'\n')
keys.clear()
print('SCOPED_CHECKS_FINISHED; only sanitized verdicts retained',flush=True)
