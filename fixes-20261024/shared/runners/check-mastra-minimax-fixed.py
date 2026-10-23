from pathlib import Path
import os,sys,json,subprocess,tempfile,time,signal,selectors,hashlib
b=Path.home()/'.ao/dev/harness-fixes-6541-6546-20261023';w=b/'mastra'
head=subprocess.check_output(['git','-C',str(w),'rev-parse','HEAD'],text=True).strip();assert head==sys.argv[1]
assert not subprocess.check_output(['git','-C',str(w),'status','--porcelain'],text=True)
print('READY_FOR_SCOPED_INPUT',flush=True)
keys=json.loads(sys.stdin.readline());assert set(keys)=={'minimax'}
node=str(Path((b/'node24-path.txt').read_text().strip())/'node')
base=b/'ephemeral-provider-checks';base.mkdir(exist_ok=True)
probe=(w/'backend/internal/adapters/agent/mastracode/auth-probe.mjs').read_text()
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

with tempfile.TemporaryDirectory(prefix='mastra-minimax-fixed-',dir=base) as td:
 d=Path(td);home=d/'home';home.mkdir();tmp=d/'tmp';tmp.mkdir();profile=d/'profile'
 env={'HOME':str(home),'PATH':str(Path(node).parent)+':/usr/bin:/bin','TMPDIR':str(tmp),'MASTRA_APP_DATA_DIR':str(profile),'MINIMAX_API_KEY':keys['minimax']}
 setup="const m=await import(process.argv[1]);const s=m.loadSettings();s.models.activeModelPackId=null;s.models.modeDefaults={...s.models.modeDefaults,build:'minimax/MiniMax-M3'};m.saveSettings(s);"
 rc,raw,reason,sec=run([node,'--input-type=module','-e',setup,'--',(b/'native/node_modules/@mastra/code-sdk/dist/onboarding/settings.js').as_uri()],env,d,10);assert rc==0
 rc,raw,reason,sec=run([node,'--input-type=module','-e',probe,'--',str(b/'native/node_modules/.bin/mastracode'),str(profile)],env,d,25)
 val=json.loads(raw);assert isinstance(val,dict) and set(val)<= {'status','code','provider','httpStatus'}
 out={'evidenceKind':'Bounded native Mastra auth-probe provider request','head':head,'authProbeSha256':hashlib.sha256(probe.encode()).hexdigest(),'provider':'minimax','model':'minimax/MiniMax-M3','status':'PASS' if val.get('status')=='authorized' else 'UNVERIFIED','exit':rc,'termination':reason,'seconds':sec,'nativeProbe':val,'credentialsRetained':False,'profileRetained':False}
 (b/'mastra-minimax-provider-after-fix.json').write_text(json.dumps(out,indent=2)+'\n');print(json.dumps(out),flush=True)
keys.clear()
