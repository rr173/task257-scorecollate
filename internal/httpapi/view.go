package httpapi

import (
	"net/http"
)

// viewHTML 是轻量校勘视图：并排展示小节与来源证据，数据全部来自后端 API。
const viewHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<title>古典乐谱异文传抄证据复核台</title>
<style>
body{font-family:-apple-system,"PingFang SC",sans-serif;max-width:960px;margin:24px auto;padding:0 16px;color:#222}
h1{font-size:20px} table{border-collapse:collapse;width:100%;margin-top:12px}
th,td{border:1px solid #ccc;padding:6px 8px;text-align:left;font-size:13px;vertical-align:top}
.badge{display:inline-block;padding:1px 8px;border-radius:10px;background:#eee;font-size:12px;margin-right:6px}
.err{background:#fdecea;color:#b00020} .ok{background:#e8f5e9;color:#1b5e20}
</style>
</head>
<body>
<h1>古典乐谱异文传抄证据复核台</h1>
<div id="app">加载中…</div>
<script>
async function j(url,opt){const r=await fetch(url,opt);if(!r.ok)throw new Error(await r.text());return r.json()}
async function main(){
  const projects=await j('/api/projects');
  if(!projects.length){document.getElementById('app').innerHTML='<p>暂无校勘项目。请先通过 API 创建项目并导入片段。</p>';return}
  const p=projects[projects.length-1];
  const [frags,variants,sources,editions]=await Promise.all([
    j('/api/projects/'+p.id+'/fragments'),
    j('/api/projects/'+p.id+'/variants'),
    j('/api/projects/'+p.id+'/sources'),
    j('/api/projects/'+p.id+'/editions')
  ]);
  let rows='';
  for(const v of variants){
    rows+='<tr><td>'+v.measure_number+'</td><td>声部'+v.voice+'</td><td>'+v.content_a+'</td><td>'+v.content_b+'</td>'+
      '<td><span class="badge">'+v.detected_kind+'</span>支持度 '+v.support_count+'</td><td>'+v.state+'</td></tr>';
  }
  document.getElementById('app').innerHTML=
    '<p><span class="badge ok">'+p.state+'</span>项目：'+(p.title||p.id)+'</p>'+
    '<p>来源 '+(sources||[]).length+' 份 · 片段 '+(frags||[]).length+' 份 · 异文 '+(variants||[]).length+' 处 · 版本 '+(editions||[]).length+' 个</p>'+
    '<h2>异文对照（小节 | 声部 | 读法 A | 读法 B | 初判 | 状态）</h2>'+
    '<table><thead><tr><th>小节</th><th>声部</th><th>参考读法</th><th>对照读法</th><th>初判/支持</th><th>状态</th></tr></thead><tbody>'+
    (rows||'<tr><td colspan="6">尚无异文。请先对齐。</td></tr>')+'</tbody></table>';
}
main().catch(e=>{document.getElementById('app').innerHTML='<p class="err">加载失败：'+e.message+'</p>'});
</script>
</body>
</html>`

func (a *API) handleView(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(viewHTML))
}
