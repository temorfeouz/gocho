package node

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"

	"github.com/donkeysharp/gocho/pkg/config"
)

const (
	HTML_BODY = `<html>
<head>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css" type="text/css">
    <style>
        * {
            font-family: sans-serif;
        }
        a {
            text-decoration: none;
            color: #1552A8;
            display: block;
            padding-bottom: 3px;
        }
        a.directory:before {
            color: #FFCF30;
            font-family: FontAwesome;
            content: "\f07b\00a0";
        }
        a.file:before {
            font-family: FontAwesome;
            content: "\f016\00a0";
		}
		 a.elem{
			 float:left;
		 }
		.actions{
			 float: right;
		}
		.container{
			clear:both;
			border-bottom: 1px solid #333;
		}
		.fa{
			cursor:pointer;
		}
		fieldset { display: inline-block }
    </style>
    <script>
        function goBack() {
            var path = window.location.pathname.split('/');
            if (path.length <= 3) {
                window.location = '/';
                return;
            }
            window.location = path.slice(0, path.length - 2).join('/');
        }
    </script>
</head>
<body>
<fieldset>
<legend>Папки</legend>
<input type="file" id="folder-input" multiple webkitdirectory allowdirs />
</fieldset>
<fieldset>
<legend>Файлы</legend>
<input type="file" id="file-input" multiple />
</fieldset>
<br>
<br>
<a class="directory" onClick="javascript:goBack()" href="#">..</a>`
	HTML_END = `



	<script type="text/javascript">
	function delelem(type, elem){
		var	params = "elem="+window.location.pathname+elem+"&delete="+type;

		var xhr = new XMLHttpRequest();
		xhr.open('POST', '/delete', true);

		xhr.onreadystatechange = function() {
			if (xhr.readyState === 4) {
				location.reload();
			}
		  }

		xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
		xhr.send(params);
	}


	    function archive(type, elem){
		        var params = "elem="+window.location.pathname+elem;
window.location='/archive?'+params;
return;
		        var xhr = new XMLHttpRequest();
		        xhr.open('GET', '/archive?'+params, true);

		        // xhr.onreadystatechange = function() {
		        //  if (xhr.readyState === 4) {
		        //      location.reload();
		        //  }
		        //   }

		        // xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
		        xhr.send(params);
		    }



    document.addEventListener('DOMContentLoaded', () => {
        document.querySelectorAll(".file-container").forEach((el) => {
            document.body.appendChild(el);
		})

		var uppie = new Uppie();

		uppie(document.querySelector('#folder-input'), function (event, formData, files) {
			var xhr = new XMLHttpRequest();
			xhr.open('POST', '/upload');

			xhr.onreadystatechange = function() {
				if (xhr.readyState === 4) {
					location.reload();
				}
			  }

			xhr.send(formData);
			})

			uppie(document.querySelector('#file-input'), function (event, formData, files) {
				var xhr = new XMLHttpRequest();
				xhr.open('POST', '/upload');

				xhr.onreadystatechange = function() {
					if (xhr.readyState === 4) {
						location.reload();
					}
				  }

				xhr.send(formData);
				})
});



	/*! uppie v1.0.9 | (c) silverwind | BSD license */
!function(e,n){"function"==typeof define&&define.amd?define([],n):"object"==typeof module&&module.exports?module.exports=n():e.Uppie=n()}("undefined"!=typeof self?self:this,function(){"use strict";return function(){return function(e,n){e instanceof NodeList?[].slice.call(e).forEach(function(e){i(e,n)}):i(e,n)}};function i(e,i){if("input"===e.tagName.toLowerCase()&&"file"===e.type)e.addEventListener("change",function(e){var n=e.target;n.files&&n.files.length?a(n,i.bind(null,e)):"getFilesAndDirectories"in n?t(n,i.bind(null,e)):i(e)});else{var n=function(e){e.preventDefault()};e.addEventListener("dragover",n),e.addEventListener("dragenter",n),e.addEventListener("drop",function(e){e.preventDefault();var n=e.dataTransfer;n.items&&n.items.length&&"webkitGetAsEntry"in n.items[0]?function(e,n){var a=new FormData,o=[],i=[];function r(e,t,i){t||(t=e.name),function i(t,e,a,o){var r=e||t.createReader();r.readEntries(function(e){var n=a?a.concat(e):e;e.length?setTimeout(i.bind(null,t,r,n,o),0):o(n)})}(e,0,0,function(e){var n=[];e.forEach(function(e){n.push(new Promise(function(i){e.isFile?e.file(function(e){var n=t+"/"+e.name;a.append("files[]",e,n),o.push(n),i()},i.bind()):r(e,t+"/"+e.name,i)}))}),Promise.all(n).then(i.bind())})}[].slice.call(e).forEach(function(e){(e=e.webkitGetAsEntry())&&i.push(new Promise(function(n){e.isFile?e.file(function(e){a.append("files[]",e,e.name),o.push(e.name),n()},n.bind()):e.isDirectory&&r(e,null,n)}))}),Promise.all(i).then(n.bind(null,a,o))}(n.items,i.bind(null,e)):"getFilesAndDirectories"in n?t(n,i.bind(null,e)):n.files?a(n,i.bind(null,e)):i()})}}function t(e,i){var o=new FormData,r=[],l=function(e,t,n){var a=[];e.forEach(function(i){a.push(new Promise(function(n){if("getFilesAndDirectories"in i)i.getFilesAndDirectories().then(function(e){l(e,i.path+"/",n)});else{if(i.name){var e=(t+i.name).replace(/^[/\\]/,"");o.append("files[]",i,e),r.push(e)}n()}}))}),Promise.all(a).then(n)};e.getFilesAndDirectories().then(function(n){new Promise(function(e){l(n,"/",e)}).then(i.bind(null,o,r))})}function a(e,n){var i=new FormData,t=[];[].slice.call(e.files).forEach(function(e){i.append("files[]",e,e.webkitRelativePath||e.name),t.push(e.webkitRelativePath||e.name)}),n(i,t)}});
</script>
</body>
</html>`
)

type FileServerResponseInterceptor struct {
	OriginalWriter http.ResponseWriter
	IndexBuffer    *bytes.Buffer
}

func (f *FileServerResponseInterceptor) WriteHeader(status int) {
	f.OriginalWriter.WriteHeader(status)
}

func (f *FileServerResponseInterceptor) Header() http.Header {
	return f.OriginalWriter.Header()
}

func (f *FileServerResponseInterceptor) Write(content []byte) (int, error) {
	// if it's not an html tag why bother evaluating with regex?
	if content[0] != byte('<') {
		return f.OriginalWriter.Write(content)
	}
	re := regexp.MustCompile("^<a.+href=\"(.+)\".*>(.+)</a>$|^</{0,1}pre>$")
	if !re.Match(bytes.Trim(content, "\n\r")) {
		return f.OriginalWriter.Write(content)
	}
	content = bytes.Trim(content, "\n\r")

	directoryRegex := regexp.MustCompile("^<a.+href=\"(.+/)\".*>(.+)</a>$")
	if directoryRegex.Match(content) {
		directoryLink := "<div class='container'><a class=\"directory elem\" href=\"$1\">$2</a>  <span class=\"actions\"><i onclick=\"archive('folder', '$1')\" class=\"fa fa-archive\"></i> <i onclick=\"delelem('folder', '$1')\" class=\"fa fa-trash\"></i></span></div>\n"

		content = directoryRegex.ReplaceAll(content, []byte(directoryLink))
		return f.IndexBuffer.Write(content)
	}
	fileRegex := regexp.MustCompile("^<a.+href=\"(.+)\".*>(.+)</a>$")
	if fileRegex.Match(content) {
		fileLink := "<div class='container file-container'><a class=\"file elem\" href=\"$1\">$2</a>   <span class=\"actions\"><i onclick=\"archive('file', '$1')\" class=\"fa fa-archive\"></i> <i onclick=\"delelem('file', '$1')\" class=\"fa fa-trash\"></i></span></div>\n"
		content = fileRegex.ReplaceAll(content, []byte(fileLink))
		return f.IndexBuffer.Write(content)
	}
	return 0, nil
}

func interceptorHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		interceptor := &FileServerResponseInterceptor{
			OriginalWriter: w,
			IndexBuffer:    bytes.NewBuffer(nil),
		}
		r.Header.Del("If-Modified-Since")
		next.ServeHTTP(interceptor, r)

		if interceptor.IndexBuffer.Len() > 0 {
			w.Write([]byte(HTML_BODY))
			w.Write(interceptor.IndexBuffer.Bytes())
			w.Write([]byte(HTML_END))
		}
	}
	return http.HandlerFunc(fn)
}

func fileServe(conf *config.Config) {
	fileMux := http.NewServeMux()
	fileMux.Handle("/", interceptorHandler(http.FileServer(http.Dir(conf.ShareDirectory))))
	fileMux.HandleFunc("/upload", fileUpload(conf))
	fileMux.HandleFunc("/delete", delete(conf))
	fileMux.HandleFunc("/archive", archive(conf))
	//uploadArchive
	http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", conf.WebPort), fileMux)
}
