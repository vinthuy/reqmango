const h=require("http");
function r(m,p,b,t){return new Promise((ok,no)=>{const o={hostname:"localhost",port:8000,path:p,method:m,headers:{"Content-Type":"application/json"}};if(t)o.headers.Authorization="Bearer "+t;const q=h.request(o,r=>{let d="";r.on("data",c=>d+=c);r.on("end",()=>{try{ok({s:r.statusCode,d:JSON.parse(d)})}catch(e){ok({s:r.statusCode,d})}})});q.on("error",no);if(b)q.write(JSON.stringify(b));q.end()})}
(async()=>{
const ts=Date.now();
await r("POST","/api/v1/auth/register",{email:"rpt"+ts+"@t.com",username:"rpt"+ts,password:"E2eTest123!",display_name:"t"});
const lg=await r("POST","/api/v1/auth/login",{email:"rpt"+ts+"@t.com",password:"E2eTest123!"});
const tk=lg.d.access_token;
const ws=await r("POST","/api/v1/workspaces",{name:"W",slug:"w"+ts},tk);
const pr=await r("POST","/api/v1/projects?workspace_id="+ws.d.id,{name:"P",identifier:"P",description:"t"},tk);
const pid=pr.d.id;
console.log("PID:",pid);
const tests=[
["dist state",{report_type:"distribution",group_by:"state",chart:"bar"}],
["RQL =",{report_type:"distribution",group_by:"state",chart:"bar",rql:'state = "Todo"'}],
["RQL !=",{report_type:"distribution",group_by:"state",chart:"bar",rql:'state != "Done"'}],
["RQL IN",{report_type:"distribution",group_by:"priority",chart:"pie",rql:'priority IN ("urgent","high")'}],
["RQL NOT IN",{report_type:"distribution",group_by:"state",chart:"bar",rql:'state NOT IN ("Done")'}],
["RQL LIKE",{report_type:"distribution",group_by:"state",chart:"bar",rql:'name LIKE "test"'}],
["RQL NOT LIKE",{report_type:"distribution",group_by:"state",chart:"bar",rql:'name NOT LIKE "debug"'}],
["RQL >=",{report_type:"distribution",group_by:"state",chart:"bar",rql:'start_date >= "2024-01-01"'}],
["RQL <=",{report_type:"distribution",group_by:"state",chart:"bar",rql:'target_date <= "2024-12-31"'}],
["RQL IS NULL",{report_type:"distribution",group_by:"state",chart:"bar",rql:'assignee IS NULL'}],
["RQL IS NOT NULL",{report_type:"distribution",group_by:"state",chart:"bar",rql:'assignee IS NOT NULL'}],
["RQL AND",{report_type:"distribution",group_by:"state",chart:"bar",rql:'priority IN ("urgent") AND state != "Done"'}],
["trend",{report_type:"created_trend",group_by:"state",chart:"line",interval:"day"}],
["vs resolved",{report_type:"created_vs_resolved",chart:"bar",interval:"week"}],
["avg_age",{report_type:"avg_age",group_by:"state",chart:"bar"}],
];
for(const[name,body] of tests){
const res=await r("POST","/api/v1/projects/"+pid+"/reports",body,tk);
const ok=res.s===200;
console.log((ok?"PASS":"FAIL")+" "+name+": "+res.s+(ok?" total:"+res.d.total:" err:"+JSON.stringify(res.d).substring(0,100)));
}
const cr=await r("POST","/api/v1/projects/"+pid+"/saved-reports",{name:"T",report_type:"distribution",group_by:"state",chart_type:"bar",rql:'priority = "high"'},tk);
console.log("saved create:",cr.s,cr.d.name);
const lr=await r("GET","/api/v1/projects/"+pid+"/saved-reports",null,tk);
console.log("saved list:",lr.s,"count:",lr.d.length);
const ur=await r("PATCH","/api/v1/projects/"+pid+"/saved-reports/"+cr.d.id,{name:"U"},tk);
console.log("saved update:",ur.s,ur.d.name);
const dr=await r("DELETE","/api/v1/projects/"+pid+"/saved-reports/"+cr.d.id,null,tk);
console.log("saved delete:",dr.s);
})().catch(e=>console.error(e.message));
