package main

const (
	mainTemplate = `<html>
	<head>
		<title>{{.Title}}</title>
		<style>
			* {
				margin: 0px;
				padding: 0px;
			}
			#navigation {
    			position: fixed;
    			height: 50px;
    			top: 0;
    			width: 100%;
    			z-index: 100;
    			background-color: #EEE;
    			-webkit-box-shadow: 0px 10px 15px 0px rgba(0,0,0,0.35);
    			-moz-box-shadow: 0px 10px 15px 0px rgba(0,0,0,0.35);
    			box-shadow: 0px 10px 15px 0px rgba(0,0,0,0.35);
			}
			#content { 
    			margin-top: 80px;
				text-align: center;
				width: 100%;
			}
		    .nav {
		    	margin-top: 10px;
				clear: both;
				display: inline-block;
				position: relative;
		    }
		    .nav a {
		    	padding: 10px;
		    	font-family: sans-serif;
		    	text-decoration: none;
		    	font-size: 25px;
		    	color: #333;
		    	text-transform: uppercase;
		    	letter-spacing: -1px;
		    }
			.thumb {
				width: {{.Size}}px;
				height: {{.Size}}px;
				background-color: #F6CECE;
				clear: both;
				display: inline-block;
				position: relative;
			}
			.thumb-inner {
				display: table;
				position: absolute; 
  				left: 10px; 
  				top: 10px; 
  				width: {{.InnerSize}}px; 
  				height: {{.InnerSize}}px; 
			}
			.thumb-inner p {
				display: table-cell; 
				vertical-align: middle; 
				text-align: center; 
				font-family: sans-serif;
				color: white;
				text-shadow: 2px 2px 2px rgba(150, 150, 150, 1);
				font-weight: bold;
			}
		</style>
	</head>
	<body>
		<div id="navigation">
		{{with .Path.Dirs}}{{range .}}
			<div class="nav"><a href="{{.Link}}">&#8226;&nbsp;&nbsp;&nbsp;{{.Name}}</a></div>
		{{end}}{{end}}</div>
		
		<div id="content">
		{{with .Links}}{{range .}}
			<a href="{{.Href}}">
				<div class='thumb' style='background-image:url("{{.Thumb}}");'>
					<div class='thumb-inner'><p>{{.Text}}</p></div></div></a>
		{{end}}{{end}}
		</div>
	</body>
</html>`
)
