(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-8bbc8b7a"],{"3b9d":function(s,o,r){},"5afe":function(s,o,r){"use strict";var e=r("3b9d"),t=r.n(e);t.a},dd7b:function(s,o,r){"use strict";r.r(o);var e=function(){var s=this,o=s.$createElement,r=s._self._c||o;return r("div",{staticClass:"login-container"},[r("div",{staticClass:"login-logo"}),r("form",{staticClass:"login-form",attrs:{name:"loginForm",novalidate:""}},[s._m(0),r("label",{staticClass:"item-input"},[r("input",{directives:[{name:"model",rawName:"v-model",value:s.loginForm.username,expression:"loginForm.username"}],staticClass:"user-login-form",attrs:{type:"text",placeholder:"用户名",autocomplete:"off",required:""},domProps:{value:s.loginForm.username},on:{focus:function(o){s.errorMsg=""},input:function(o){o.target.composing||s.$set(s.loginForm,"username",o.target.value)}}})]),r("label",{staticClass:"item-input type-item"},["checkbox"===(s.passwordType?"password":"text")?r("input",{directives:[{name:"model",rawName:"v-model",value:s.loginForm.password,expression:"loginForm.password"}],staticClass:"user-login-form",class:{"input-paw":s.loginForm.password,"input-paw-open":!s.passwordType},attrs:{placeholder:"密码",autocomplete:"off",required:"",type:"checkbox"},domProps:{checked:Array.isArray(s.loginForm.password)?s._i(s.loginForm.password,null)>-1:s.loginForm.password},on:{focus:function(o){s.errorMsg=""},change:function(o){var r=s.loginForm.password,e=o.target,t=!!e.checked;if(Array.isArray(r)){var a=null,n=s._i(r,a);e.checked?n<0&&s.$set(s.loginForm,"password",r.concat([a])):n>-1&&s.$set(s.loginForm,"password",r.slice(0,n).concat(r.slice(n+1)))}else s.$set(s.loginForm,"password",t)}}}):"radio"===(s.passwordType?"password":"text")?r("input",{directives:[{name:"model",rawName:"v-model",value:s.loginForm.password,expression:"loginForm.password"}],staticClass:"user-login-form",class:{"input-paw":s.loginForm.password,"input-paw-open":!s.passwordType},attrs:{placeholder:"密码",autocomplete:"off",required:"",type:"radio"},domProps:{checked:s._q(s.loginForm.password,null)},on:{focus:function(o){s.errorMsg=""},change:function(o){return s.$set(s.loginForm,"password",null)}}}):r("input",{directives:[{name:"model",rawName:"v-model",value:s.loginForm.password,expression:"loginForm.password"}],staticClass:"user-login-form",class:{"input-paw":s.loginForm.password,"input-paw-open":!s.passwordType},attrs:{placeholder:"密码",autocomplete:"off",required:"",type:s.passwordType?"password":"text"},domProps:{value:s.loginForm.password},on:{focus:function(o){s.errorMsg=""},input:function(o){o.target.composing||s.$set(s.loginForm,"password",o.target.value)}}}),s.loginForm.password?r("button",{staticClass:"paw-btn",class:{"paw-btn-open":!s.passwordType},attrs:{type:"button"},on:{click:function(o){s.passwordType=!s.passwordType}}}):s._e()]),r("p",{staticClass:"error-message"},[s._v("\n            "+s._s(s.errorMsg)+"\n        ")]),r("button",{staticClass:"button",on:{click:function(o){return o.preventDefault(),s.submit(o)}}},[s._v("登录")]),r("div",{staticClass:"mark"})])])},t=[function(){var s=this,o=s.$createElement,e=s._self._c||o;return e("div",{staticClass:"login-title"},[e("img",{staticStyle:{height:"23px"},attrs:{src:r("84d7")}}),e("h4",[s._v("智能网关")])])}],a=(r("96cf"),r("3b8d")),n={layout:"login",data:function(){return{loginForm:{username:"",password:""},errorMsg:"",passwordType:!0}},methods:{submit:function(){var s=Object(a["a"])(regeneratorRuntime.mark((function s(){return regeneratorRuntime.wrap((function(s){while(1)switch(s.prev=s.next){case 0:if(this.loginForm.username&&this.loginForm.password){s.next=3;break}return this.errorMsg="用户名或密码不能为空！",s.abrupt("return");case 3:return this.errorMsg="",s.next=6,this.$store.dispatch("user/LoginByUsername",{username:this.loginForm.username,password:this.loginForm.password});case 6:this.$router.push({path:"/input/index"});case 7:case"end":return s.stop()}}),s,this)})));function o(){return s.apply(this,arguments)}return o}()}},i=n,l=(r("5afe"),r("2877")),p=Object(l["a"])(i,e,t,!1,null,"66f43b44",null);o["default"]=p.exports}}]);
//# sourceMappingURL=chunk-8bbc8b7a.e05b6d56.js.map