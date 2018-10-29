webpackJsonp([13],{"9K8c":function(e,t,r){"use strict";var a=function(){var e=this,t=e.$createElement,r=e._self._c||t;return r("div",{staticClass:"app-container"},[r("h1",[e._v("\n        主图列表\n    ")]),r("el-select",{staticClass:"mb20",attrs:{placeholder:"请选择"},model:{value:e.currentSystem,callback:function(t){e.currentSystem=t},expression:"currentSystem"}},e._l(e.systems,function(e){return r("el-option",{key:e.key,attrs:{label:e.label,value:e.key}})})),r("el-table",{attrs:{data:e.filterTableData,stripe:!0,size:"small",border:""}},[r("el-table-column",{attrs:{type:"index"}}),r("el-table-column",{attrs:{prop:"key",label:"Key",width:"200"}}),r("el-table-column",{attrs:{prop:"description",label:"描述"}}),r("el-table-column",{attrs:{prop:"created_at",label:"创建时间"}}),r("el-table-column",{attrs:{prop:"updated_at",label:"更新时间"}}),r("el-table-column",{attrs:{label:"操作",fixed:"right",width:"100"},scopedSlots:e._u([{key:"default",fn:function(t){return[r("el-button",{attrs:{type:"primary",size:"small",round:""},on:{click:function(r){e.editKey(t.row.key)}}},[e._v("编辑")])]}}])})],1),r("el-dialog",{attrs:{title:"添加Key",visible:e.isModalShow,width:"500px"},on:{"update:visible":function(t){e.isModalShow=t}}},[r("el-form",{attrs:{model:e.form}},[r("el-form-item",{attrs:{label:"Key","label-width":e.formLabelWidth}},[r("el-input",{attrs:{autoComplete:"off"},model:{value:e.form.key,callback:function(t){e.$set(e.form,"key",t)},expression:"form.key"}})],1),r("el-form-item",{attrs:{label:"描述","label-width":e.formLabelWidth}},[r("el-input",{attrs:{autoComplete:"off"},model:{value:e.form.description,callback:function(t){e.$set(e.form,"description",t)},expression:"form.description"}})],1)],1),r("div",{staticClass:"dialog-footer",attrs:{slot:"footer"},slot:"footer"},[r("el-button",{on:{click:function(t){e.isModalShow=!1}}},[e._v("取 消")]),r("el-button",{attrs:{type:"primary"},on:{click:e.createKey}},[e._v("确 定")])],1)],1)],1)};a._withStripped=!0;var n={render:a,staticRenderFns:[]};t.a=n},MCja:function(e,t,r){"use strict";var a=r("Xxa5"),n=r.n(a),o=r("exGp"),s=r.n(o);t.a={name:"diagram",data:function(){return{tableData:[],isModalShow:!1,form:{key:"",description:""},formLabelWidth:"60px",currentKey:"",currentSystem:"topo",systems:[{label:"所有系统",key:"topo"},{label:"通风系统",key:"vent"},{label:"电力监控",key:"power"},{label:"皮带运输",key:"belt"},{label:"压风系统",key:"compress"}]}},computed:{filterTableData:function(){var e=this;return this.tableData.filter(function(t){return t.key.indexOf("topo")>=0&&t.key.indexOf(e.currentSystem)>=0})}},methods:{getData:function(){var e=s()(n.a.mark(function e(){var t;return n.a.wrap(function(e){for(;;)switch(e.prev=e.next){case 0:return e.next=2,this.$axios({method:"get",url:"/config-center/list-keys",headers:{Token:"Mercury"}});case 2:t=e.sent,this.tableData=t.data;case 4:case"end":return e.stop()}},e,this)}));return function(){return e.apply(this,arguments)}}(),createKey:function(){var e=s()(n.a.mark(function e(){var t,r;return n.a.wrap(function(e){for(;;)switch(e.prev=e.next){case 0:if(e.prev=0,t={key:this.form.key,value:{}},this.form.key){e.next=5;break}return this.$message({message:"Key不能为空",type:"error"}),e.abrupt("return");case 5:return e.next=7,this.$axios({method:"post",url:"config-center/config",headers:{Token:"Mercury"},data:t});case 7:r=e.sent,console.log(r),this.tableData.push(r.data),this.isModalShow=!1,e.next=16;break;case 13:e.prev=13,e.t0=e.catch(0),this.$message({message:e.t0.response?e.t0.response.data:e.t0,type:"error"});case 16:case"end":return e.stop()}},e,this,[[0,13]])}));return function(){return e.apply(this,arguments)}}(),deleteKey:function(){var e=s()(n.a.mark(function e(){var t=this;return n.a.wrap(function(e){for(;;)switch(e.prev=e.next){case 0:return e.prev=0,e.next=3,this.$axios({method:"delete",url:"config-center/config/"+this.currentKey,headers:{Token:"Mercury"}});case 3:e.sent,this.tableData=this.tableData.filter(function(e){return e.key!==t.currentKey}),e.next=10;break;case 7:e.prev=7,e.t0=e.catch(0),this.$message({message:e.t0.response.data,type:"error"});case 10:case"end":return e.stop()}},e,this,[[0,7]])}));return function(){return e.apply(this,arguments)}}(),confirmDelete:function(e){var t=this;this.currentKey=e,this.$confirm("确认要删除key： "+this.currentKey+"?","提示",{confirmButtonText:"确定",cancelButtonText:"取消",type:"warning"}).then(function(){t.deleteKey()})},editKey:function(e){this.currentKey=e,this.$router.push("/frontend/topo/"+e)}},mounted:function(){this.getData()}}},m5bp:function(e,t,r){"use strict";Object.defineProperty(t,"__esModule",{value:!0});var a=r("MCja"),n=r("9K8c"),o=r("VU/8")(a.a,n.a,!1,null,null,null);o.options.__file="pages/frontend/topo/index.vue",t.default=o.exports}});