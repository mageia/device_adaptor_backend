webpackJsonp([11],{"/PK/":function(t,e,n){"use strict";var s=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("div",{staticClass:"app-container"},[n("h1",{staticClass:"page-title"},[t._v("安全监控配置")]),n("el-row",[n("el-col",{attrs:{span:20}},[n("el-input",{staticClass:"mb20",attrs:{type:"textarea",rows:20,placeholder:"请输入内容"},model:{value:t.securityJson,callback:function(e){t.securityJson=e},expression:"securityJson"}}),n("el-button",{staticClass:"pull-right",attrs:{type:"primary"},on:{click:t.updateJson}},[t._v("保存修改")])],1)],1)],1)};s._withStripped=!0;var r={render:s,staticRenderFns:[]};e.a=r},Qi69:function(t,e,n){var s=n("nBXC");"string"==typeof s&&(s=[[t.i,s,""]]),s.locals&&(t.exports=s.locals);n("rjj0")("57027ab8",s,!1,{sourceMap:!1})},bY3o:function(t,e,n){"use strict";Object.defineProperty(e,"__esModule",{value:!0});var s=n("ozwg"),r=n("/PK/"),a=!1;var o=function(t){a||n("Qi69")},i=n("VU/8")(s.a,r.a,!1,o,null,null);i.options.__file="pages/backend/security.vue",e.default=i.exports},nBXC:function(t,e,n){(t.exports=n("FZ+f")(!1)).push([t.i,".json-textarea{width:100%;height:400px;padding:20px}",""])},ozwg:function(t,e,n){"use strict";var s=n("Xxa5"),r=n.n(s),a=n("exGp"),o=n.n(a);e.a={layout:"default",components:{},data:function(){return{securityJson:""}},computed:{},watch:{},methods:{fetchJson:function(){var t=o()(r.a.mark(function t(){var e;return r.a.wrap(function(t){for(;;)switch(t.prev=t.next){case 0:return t.next=2,this.$axios.$get("/config/videoConfig");case 2:e=t.sent,console.log(e);case 4:case"end":return t.stop()}},t,this)}));return function(){return t.apply(this,arguments)}}(),updateJson:function(){var t=o()(r.a.mark(function t(){var e,n;return r.a.wrap(function(t){for(;;)switch(t.prev=t.next){case 0:return e={key:"videoConfig",value:this.securityJson},console.log(e),t.next=4,this.$axios.$put("/config/videoConfig",e);case 4:n=t.sent,console.log(n);case 6:case"end":return t.stop()}},t,this)}));return function(){return t.apply(this,arguments)}}(),fetchSomething:function(){var t=o()(r.a.mark(function t(){var e,n;return r.a.wrap(function(t){for(;;)switch(t.prev=t.next){case 0:return e="rtsp://sample-rtsp:554/live.h264",t.next=3,this.$axios.$post("/retainTask",{url:e});case 3:n=t.sent,console.log(n);case 5:case"end":return t.stop()}},t,this)}));return function(){return t.apply(this,arguments)}}()},mounted:function(){this.fetchSomething(),this.fetchJson()}}}});