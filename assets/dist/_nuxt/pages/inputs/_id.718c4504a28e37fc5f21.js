webpackJsonp([6],{"3VE0":function(t,a,e){(t.exports=e("FZ+f")(!1)).push([t.i,".json-editor[data-v-4c35b866]{margin-bottom:15px;height:500px}",""])},QSqw:function(t,a,e){"use strict";var i=e("Xxa5"),n=e.n(i),r=e("exGp"),l=e.n(r);a.a={name:"KeyEdit",components:{},data:function(){return{formLabelWidth:"80px",units:["s","ms"],id:"",data:{}}},computed:{inputType:function(){return JSON.parse(window.localStorage.getItem("INPUT_PLUGIN_TYPE"))}},methods:{fetchData:function(){var t=l()(n.a.mark(function t(){var a,e;return n.a.wrap(function(t){for(;;)switch(t.prev=t.next){case 0:return t.next=2,this.$axios({method:"get",url:"/plugin/inputs/"+this.id});case 2:a=t.sent,e=a.data,this.data=e,e.interval&&(e.unit=/[a-z]*$/.exec(e.interval)[0],e.interval=/^[0-9]*/.exec(e.interval)[0]);case 6:case"end":return t.stop()}},t,this)}));return function(){return t.apply(this,arguments)}}(),updateDataSource:function(){var t=l()(n.a.mark(function t(){var a;return n.a.wrap(function(t){for(;;)switch(t.prev=t.next){case 0:return a={name_override:this.data.name_override,plugin_name:this.data.plugin_name,interval:""+this.data.interval+this.data.unit,point_map:this.data.point_map},t.next=3,this.$axios({method:"put",url:"/plugin/inputs/"+this.id,data:a});case 3:t.sent,this.$message({message:"修改成功！",type:"success"}),this.isEditModalShow=!1;case 6:case"end":return t.stop()}},t,this)}));return function(){return t.apply(this,arguments)}}(),jumpTo:function(){this.$router.push("/")}},mounted:function(){this.$nextTick(function(){this.id=this.$route.params.id,this.fetchData(this.key)})}}},fyRH:function(t,a,e){"use strict";Object.defineProperty(a,"__esModule",{value:!0});var i=e("QSqw"),n=e("jwlg"),r=!1;var l=function(t){r||e("mU4Y")},s=e("VU/8")(i.a,n.a,!1,l,"data-v-4c35b866",null);s.options.__file="pages/inputs/_id.vue",a.default=s.exports},jwlg:function(t,a,e){"use strict";var i=function(){var t=this,a=t.$createElement,e=t._self._c||a;return e("div",{staticClass:"app-container",staticStyle:{width:"900px",margin:"0 auto"}},[e("h1",[t._v("编辑数据接入数据源")]),e("el-card",{staticClass:"mb20",staticStyle:{background:"#f5f6f7"},attrs:{shadow:"never"}},[e("el-form",{attrs:{model:t.data}},[e("el-form-item",{attrs:{label:"名称","label-width":t.formLabelWidth}},[e("el-input",{attrs:{autoComplete:"off"},model:{value:t.data.name_override,callback:function(a){t.$set(t.data,"name_override",a)},expression:"data.name_override"}})],1),e("el-form-item",{attrs:{label:"类型","label-width":t.formLabelWidth}},[e("el-select",{attrs:{placeholder:"选择数据源类型"},model:{value:t.data.plugin_name,callback:function(a){t.$set(t.data,"plugin_name",a)},expression:"data.plugin_name"}},t._l(t.inputType,function(t){return e("el-option",{key:t,attrs:{label:t,value:t}})}))],1),e("el-form-item",{attrs:{label:"采集频率","label-width":t.formLabelWidth}},[e("el-input",{staticStyle:{width:"80px","margin-right":"10px"},attrs:{autoComplete:"off"},model:{value:t.data.interval,callback:function(a){t.$set(t.data,"interval",a)},expression:"data.interval"}}),e("el-select",{staticStyle:{width:"70px"},model:{value:t.data.unit,callback:function(a){t.$set(t.data,"unit",a)},expression:"data.unit"}},t._l(t.units,function(t){return e("el-option",{key:t,attrs:{label:t,value:t}})}))],1),e("el-form-item",{attrs:{label:"地址","label-width":t.formLabelWidth}},[e("el-input",{attrs:{autoComplete:"off"},model:{value:t.data.point_map_path,callback:function(a){t.$set(t.data,"point_map_path",a)},expression:"data.point_map_path"}})],1),e("el-form-item",{attrs:{label:"点表","label-width":t.formLabelWidth}},[e("el-input",{attrs:{type:"textarea",autoComplete:"off",rows:10},model:{value:t.data.point_map_content,callback:function(a){t.$set(t.data,"point_map_content",a)},expression:"data.point_map_content"}})],1)],1)],1),e("el-button",{staticClass:"pull-right",attrs:{type:"primary",size:"small"},on:{click:function(a){t.updateDataSource()}}},[t._v("保存")]),e("el-button",{staticClass:"pull-right",staticStyle:{"margin-right":"20px"},attrs:{type:"default",size:"small"},on:{click:function(a){t.jumpTo()}}},[t._v("返回上一页")])],1)};i._withStripped=!0;var n={render:i,staticRenderFns:[]};a.a=n},mU4Y:function(t,a,e){var i=e("3VE0");"string"==typeof i&&(i=[[t.i,i,""]]),i.locals&&(t.exports=i.locals);e("rjj0")("a64a2c1a",i,!1,{sourceMap:!1})}});