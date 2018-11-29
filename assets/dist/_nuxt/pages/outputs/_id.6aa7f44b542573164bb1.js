webpackJsonp([1],{"97Cl":function(t,e,a){"use strict";var n=a("ueV6"),i=a("GQpR"),o=a("VU/8")(n.a,i.a,!1,null,null,null);o.options.__file="components/lean-form.vue",e.a=o.exports},"9UqO":function(t,e,a){"use strict";Object.defineProperty(e,"__esModule",{value:!0});var n=a("q9dy"),i=a("DeVu"),o=!1;var r=function(t){o||a("Z5dU")},s=a("VU/8")(n.a,i.a,!1,r,"data-v-abe4d158",null);s.options.__file="pages/outputs/_id.vue",e.default=s.exports},DeVu:function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"app-container",staticStyle:{width:"900px",margin:"0 auto"}},[a("h1",[t._v("编辑数据源")]),a("el-card",{staticClass:"mb20"},[a("lean-form",{attrs:{"sample-data":t.outputConfig,"form-data":t.data,"is-editing":!0,"current-type":t.data.plugin_name},on:{changeFormData:t.changeFormData}})],1),a("el-button",{staticClass:"pull-right",attrs:{type:"primary",size:"small"},on:{click:function(e){t.updateDataSource()}}},[t._v("保存")]),a("el-button",{staticClass:"pull-right",staticStyle:{"margin-right":"20px"},attrs:{type:"default",size:"small"},on:{click:function(e){t.jumpTo()}}},[t._v("返回上一页")])],1)};n._withStripped=!0;var i={render:n,staticRenderFns:[]};e.a=i},GQpR:function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",[a("el-form",{ref:"leanForm",attrs:{model:t.form}},[t.isEditing?a("el-form-item",{attrs:{label:"类型","label-width":t.formLabelWidth}},[a("span",[t._v(t._s(t.currentType))])]):t._e(),t.isEditing?t._e():a("el-form-item",{attrs:{label:"类型","label-width":t.formLabelWidth}},[a("el-select",{attrs:{placeholder:"选择数据源类型"},model:{value:t.form.plugin_name,callback:function(e){t.$set(t.form,"plugin_name",e)},expression:"form.plugin_name"}},t._l(t.inputType,function(t){return a("el-option",{key:t,attrs:{label:t,value:t}})}))],1),t._l(t.form.fields,function(e,n){return a("el-form-item",{key:n,attrs:{label:e.label,"label-width":t.formLabelWidth}},["input"===e.type?a("el-input",{attrs:{autoComplete:"off"},model:{value:e.value,callback:function(a){t.$set(e,"value",a)},expression:"item.value"}}):t._e(),"combine"===e.type?[a("el-input",{staticStyle:{width:"80px","margin-right":"10px"},attrs:{autoComplete:"off"},model:{value:e.value,callback:function(a){t.$set(e,"value",a)},expression:"item.value"}}),a("el-select",{staticStyle:{width:"70px"},model:{value:e.unit,callback:function(a){t.$set(e,"unit",a)},expression:"item.unit"}},t._l(t.units,function(t){return a("el-option",{key:t,attrs:{label:t,value:t}})}))]:t._e(),"radio"===e.type?[a("el-radio-group",{model:{value:e.value,callback:function(a){t.$set(e,"value",a)},expression:"item.value"}},[a("el-radio",{attrs:{label:!0}},[t._v("true")]),a("el-radio",{attrs:{label:!1}},[t._v("false")])],1)]:t._e(),"text"===e.type?a("el-input",{attrs:{type:"textarea",autoComplete:"off",rows:10},model:{value:e.value,callback:function(a){t.$set(e,"value",a)},expression:"item.value"}}):t._e()],2)})],2)],1)};n._withStripped=!0;var i={render:n,staticRenderFns:[]};e.a=i},K8Rp:function(t,e,a){(t.exports=a("FZ+f")(!1)).push([t.i,".json-editor[data-v-abe4d158]{margin-bottom:15px;height:500px}",""])},Z5dU:function(t,e,a){var n=a("K8Rp");"string"==typeof n&&(n=[[t.i,n,""]]),n.locals&&(t.exports=n.locals);a("rjj0")("703111b5",n,!1,{sourceMap:!1})},q9dy:function(t,e,a){"use strict";var n=a("woOf"),i=a.n(n),o=a("Xxa5"),r=a.n(o),s=a("exGp"),u=a.n(s),l=a("97Cl");e.a={name:"KeyEdit",components:{LeanForm:l.a},data:function(){return{id:"",data:{},outputConfig:{},currentData:{}}},computed:{outputType:function(){return JSON.parse(window.localStorage.getItem("OUTPUT_PLUGIN_TYPE"))}},methods:{fetchData:function(){var t=u()(r.a.mark(function t(){var e,a;return r.a.wrap(function(t){for(;;)switch(t.prev=t.next){case 0:return t.next=2,this.$axios({method:"get",url:"/plugin/outputs/"+this.id});case 2:e=t.sent,a=e.data,this.data=a;case 5:case"end":return t.stop()}},t,this)}));return function(){return t.apply(this,arguments)}}(),updateDataSource:function(){var t=u()(r.a.mark(function t(){var e;return r.a.wrap(function(t){for(;;)switch(t.prev=t.next){case 0:return e={plugin_name:this.data.plugin_name},i()([],this.currentData.fields).forEach(function(t){"combine"!==t.type?e[t.key]=t.value:e[t.key]=""+t.value+t.unit}),t.next=5,this.$axios({method:"put",url:"/plugin/outputs/"+this.id,data:e});case 5:t.sent,this.$message({message:"修改成功！",type:"success"}),this.isEditModalShow=!1;case 8:case"end":return t.stop()}},t,this)}));return function(){return t.apply(this,arguments)}}(),jumpTo:function(){this.$router.push("/outputs")},initConfig:function(){this.outputConfig=JSON.parse(localStorage.getItem("OUTPUT_PLUGIN_CONFIG"))},changeFormData:function(t){this.currentData=t}},mounted:function(){this.$nextTick(function(){this.id=this.$route.params.id,this.fetchData(this.key),this.initConfig()})}}},ueV6:function(t,e,a){"use strict";var n=a("fZjL"),i=a.n(n);e.a={name:"lean-form",props:{sampleData:{type:Object,default:function(){return{}}},reset:{type:Boolean,default:!1},isEditing:{type:Boolean,default:!1},formData:{type:Object,default:function(){return{}}},currentType:{type:String,default:""}},data:function(){return{formLabelWidth:"140px",form:{plugin_name:"",fields:[]},units:["s","ms"]}},computed:{dataType:function(){return console.log(this.form.plugin_name),this.form.plugin_name},inputType:function(){return i()(this.sampleData)}},watch:{currentType:function(t,e){console.log(this.formData),this.formatForm()},dataType:function(t,e){this.formatForm()},"form.fields":{handler:function(){console.log(this.form),this.$emit("changeFormData",this.form)},deep:!0},reset:function(t,e){this.resetForm()}},methods:{formatForm:function(){var t=this;this.form.fields=[];var e=this.isEditing?this.currentType:this.dataType;e&&i()(this.sampleData).length>0&&this.sampleData[e].forEach(function(e,a){if("none"!==e.Type&&"plugin_name"!==e.Key)if("combine"===e.Type){var n=t.isEditing?t.formData[e.Key]:e.Default,i=/[a-z]*$/.exec(n)[0],o=/^[0-9]*/.exec(n)[0];t.form.fields.push({key:e.Key,value:o,label:e.Label,type:e.Type,unit:i})}else{if("text"===e.Type&&!t.isEditing)return;t.form.fields.push({key:e.Key,value:t.isEditing?t.formData[e.Key]:e.Default,label:e.Label,type:e.Type})}})},formChange:function(){console.log("change")},resetForm:function(){this.form={plugin_name:"",fields:[]}}},mounted:function(){var t=this;this.$nextTick(function(){t.formatForm()})}}}});