var slippymap = slippymap || {};

slippymap.crosshairs = (function(){

    var latlon = true;

    var self = {

	'init': function(map){

	    var container = map.getContainer();
	    var id = container.id;

	    var draw = function(){
		self.draw_crosshairs(id);
	    };

	    window.onresize = draw;

	    var coords = function(){
		self.draw_coords(map);
	    };

	    map.on('move', coords);
	    map.on('dragend', coords);
	    map.on('zoomend', coords);
	
	    // because for SOME REASON these don't both work reliably in map.on('load')
	    // because... COMPUTERS? (20160809/thisisaaronland)

	    draw();
	    coords();
	},

	'draw_coords': function(map){

	    var coords = document.getElementById("slippymap-coords");

	    if (! coords){

		var coords = document.createElement("div");
		coords.setAttribute("id", "slippymap-coords");

		var container = map.getContainer();
		var container_el = document.getElementById(container.id);

		container_el.parentNode.insertBefore(coords, container_el.nextSibling); 
	    }

	    coords.onclick = function(){
		latlon = (latlon) ? false : true;
		self.draw_coords(map);
		return;
	    };

	    var pos = map.getCenter();
	    var lat = pos['lat'];
	    var lon = pos['lng'];	  
	    
	    var zoom = map.getZoom();

	    var ll = undefined;
	    var title = undefined;

	    if (latlon){

		    ll = lat.toFixed(6) + ", " + lon.toFixed(6) + " #" + zoom.toFixed(2);
		title = "coordinates are displayed as latitude,longitude – click to toggle";
	    }
	    
	    else {

		ll = lon.toFixed(6) + ", " + lat.toFixed(6) + " #" + zoom;
		title = "coordinates are displayed as longitude,latitude – click to toggle";
	    }

	    coords.setAttribute("title", title);
	    coords.innerText = ll;	    
	},
	
	'draw_crosshairs': function(id){

	    var m = document.getElementById(id);

	    if (! m){
		return false;
	    }

	    var container = m.getBoundingClientRect();
	    
	    var height = container.height;
	    var width = container.width;
	    
	    var crosshair_y = (height / 2) - 8;
	    var crosshair_x = (width / 2);
	    
	    // http://www.sveinbjorn.org/dataurls_css
	    
	    var data_url = '"data:image/gif;base64,R0lGODlhEwATAKEBAAAAAP///////////' + 
		'yH5BAEKAAIALAAAAAATABMAAAIwlI+pGhALRXuRuWopPnOj7hngEpRm6Z' + 
		'ymAbTuC7eiitJlNHr5tmN99cNdQpIhsVIAADs="';
	    
	    var style = [];
	    style.push("position:absolute");
	    style.push("top:" + crosshair_y + "px");
	    style.push("height:19px");
	    style.push("width:19px");
	    style.push("left:" + crosshair_x + "px");
	    style.push("margin-left:-8px;");
	    style.push("display:block");
	    style.push("background-position: center center");
	    style.push("background-repeat: no-repeat");
	    style.push("background: url(" + data_url + ")");
	    style.push("z-index:10000");
	    
	    style = style.join(";");

	    var crosshairs = document.getElementById("slippymap-crosshairs");

	    if (! crosshairs){

		crosshairs = document.createElement("div");
		crosshairs.setAttribute("id", "slippymap-crosshairs");
		m.appendChild(crosshairs);
	    }

	    crosshairs.style.cssText = style;
	    return true;
	},
    };

    return self;

})();
