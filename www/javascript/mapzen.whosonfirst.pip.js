window.addEventListener("load", function load(event){

	var map;
	var marker;			    
	var candidates;
	var intersecting = [];

	var jump_to = function(str_latlon){

		str_latlon = str_latlon.trim();
		
		if (str_latlon == ""){
			return false;
		}
		
		var latlon = str_latlon.split(",");

		if (latlon.length != 2) {
			alert("Invalid lat,lon pair");
			return false;
		}

		var lat = parseFloat(latlon[0]);

		if (! lat){
			alert("Invalid latitude");
			return false;
		}

		var lon = parseFloat(latlon[1]);
		
		if (! lon){
			alert("Invalid longitude");
			return false;
		}

		map.setView([ lat, lon ], map.getZoom());

		draw_coords(lat, lon);		
	};
	
	var draw_coords = function(lat, lon){

		if ((! lat) || (! lon)){
			var center = map.getCenter();
			lat = center.lat;
			lon = center.lng;
		}
		
		lat = lat.toFixed(6);
		lon = lon.toFixed(6);
				
		var geojson = {
			"type": "Feature",
			"geometry": { "type": "Point", "coordinates": [ lon, lat ] }
		};
		
		if (marker){
			marker.remove(map);
		}
		
		marker = L.geoJSON(geojson);
		marker.addTo(map);
	};
	
	var fetch = function(url, cb){
		
		var req = new XMLHttpRequest();
		
		req.onload = function(){
			
			var rsp;
			
			try {
				rsp = JSON.parse(this.responseText);
            		}
			
			catch (e){
				console.log("ERR", url, e);
				return false;
			}
			
			cb(rsp);
       		};
		
		req.open("get", url, true);
		req.send();
	}
	
	var fetch_candidates = function(lat, lon){

		// the /candidates endpoint returns geojson by default
		
		var url = 'http://' + location.host + '/candidates?latitude=' + lat + '&longitude=' + lon;
		
		var onsuccess = function(rsp){
			
			if (candidates){
				candidates.remove(map);
			}
			
			var oneach = function(f, l){
				var props = f["properties"];
				l.bindPopup(props["id"]);
			};
			
			var args = {
				onEachFeature: oneach,						       
			};
			
			candidates = L.geoJSON(rsp, args);
			candidates.addTo(map);
			
			var candidates_list = document.getElementById("candidates-list");
			candidates_list.innerHTML = "";
			
			var features = rsp["features"];
			var count_features = features.length;
			
			for (var i=0; i < count_features; i++){
				
				var props = features[i]["properties"];
				var id = props["id"];

				var code = document.createElement("code")
				code.appendChild(document.createTextNode(id));
				
				var item = document.createElement("li");
				item.setAttribute("id", "candidate-" + id);
				
				item.appendChild(code);
				candidates_list.appendChild(item);
			}
       		};
		
		fetch(url, onsuccess);
	};
	
	var fetch_intersecting = function(lat, lon){
		
		var count_intersecting = intersecting.length;
		
		for (var i=0; i < count_intersecting; i++){
			intersecting[i].remove(map);
		}
		
		intersecting = [];
		
		var url = 'http://' + location.host + '/?latitude=' + lat + '&longitude=' + lon + '&format=geojson';
		
		var onsuccess = function(rsp){

			if ((rsp["type"]) && ((rsp["type"] == "FeatureCollection") || (rsp["type"] == "Feature"))){
				show_geojson(rsp);
				return;
			}		
			
			var places = rsp["places"];
			var count = places.length;						    

			for (var i=0; i < count; i++){
				
				var spr = places[i];
				var url = spr["mz:uri"];

				if (! url){
					console.log("missing mz:uri property, so skipping");
					return;
				}

				// update_candidate(spr);
				fetch_geojson(url);
			}
       		};
		
		fetch(url, onsuccess);
	};

	var update_candidate = function(props) {

	    var id = props["wof:id"];
	    var name = props["wof:name"];
				
	    var c = document.getElementById("candidate-" + id);

	    if (!c){
		console.log("UNABLE TO UPDATE CANDIDATE", id);
		return;
	    }

	    c.appendChild(document.createTextNode(" " + name));
	    c.setAttribute("class", "intersects");
	};

	var show_geojson = function(rsp){

		var style = {
			"color": "#FF69B4",
			"weight": 5,
			"opacity": 0.85
		};
		
		var oneach = function(f, l){
			var props = f["properties"];
			l.bindPopup(props["wof:name"]);
		};
		
		var args = {
			style: style,
			onEachFeature: oneach,						       
		};
		
		var layer = L.geoJSON(rsp, args);
		layer.addTo(map);
		
		intersecting.push(layer);

		var features = rsp["features"];
		var count = features.length;

		for (var i=0; i < count; i++){
		    var feature = features[i];
		    var props = feature["properties"];
		    update_candidate(props);
		}
	};
	
	var fetch_geojson = function(url){
		
		var onsuccess = function(rsp){
			show_geojson(rsp);
       		};
		
		fetch(url, onsuccess);
	};
	
	var pip = function(){
		
		var center = map.getCenter();
		var lat = center.lat;
		var lon = center.lng;
		
		lat = lat.toFixed(6);
		lon = lon.toFixed(6);
		
		fetch_candidates(lat, lon);
		fetch_intersecting(lat, lon);				
	};
	
	var api_key = document.body.getAttribute("data-mapzen-api-key");

	var sources = {
	    mapzen: {
		url: 'https://{s}.tile.nextzen.org/tilezen/vector/v1/512/all/{z}/{x}/{y}.topojson',
		url_subdomains: ['a', 'b', 'c', 'd'],
		url_params: {
		    api_key: api_key	// not clear this actually works... ?
		},
		tile_size: 512,
		max_zoom: 15
	    }
	};

	var scene = {
                import: [
			 "/tangram/refill-style.zip",
			 "/tangram/refill-style-themes-label.zip",
			 ],
                sources: sources,
		global: {
		    sdk_mapzen_api_key: api_key,
		},
	};

	var attributions = {
	    "Tangram": "https://github.com/tangrams/",
	    "© OSM contributors": "http://www.openstreetmap.org/",
	    "Who\"s On First": "http://www.whosonfirst.org/",
	    "Nextzen": "https://nextzen.org/",
	};

	var attrs = [];

	for (var label in attributions){

	    var link = attrs[label];

	    if (! link){
		attrs.push(label);
		continue;
	    }
	    
	    var anchor = '<a href="' + link + '" target="_blank">' + enc_label + '</a>';
	    attrs.push(anchor);
	}

	var str_attributions = attrs.join(" | ");

	// waiting for nextzen.js to be updated to do all the things - that said it's
	// not entirely clear we need all of (map/next)zen.js and could probably get
	// away with leaflet + tangram but for now we'll just keep on as-is...
	// (20180304/thisisaaronland)

	/*

	L.Mapzen.apiKey = api_key;

	var map_opts = {
	    tangramOptions: {
		scene: scene,
		attribution: attributions,
	    }
	};
	
	map = L.Mapzen.map('map', map_opts);

        var layers = [
		"neighbourhood",			    
                "locality",
		"region",
		"country"
        ];
	
        var params = {
		"sources": "wof"
        };
	
	var opts = {
		"layers": layers,
		"params": params,
	};
	
	var geocoder = L.Mapzen.geocoder(opts);
	geocoder.addTo(map);
	
	L.Mapzen.hash({
		map: map
	});
	*/

	// hack pending updates to nextzen/mapzen.js - or someone tells me
	// what the correct syntax for doing the equivalent of this is...
	// (20180304/thisisaaronland)

	// note that we assume Leaflet has been loaded by mapzen.js and we
	// explicitly load tangram.js in index.html
	// (20180304/thisisaaronland)

	map = L.map("map");

	var tangram = Tangram.leafletLayer({
		scene: scene,
		numWorkers: 2,
		unloadInvisibleTiles: false,
		updateWhenIdle: false,
		attribution: str_attributions,
	    });

	tangram.addTo(map);

	// END OF hack pending updates to nextzen/mapzen.js

	// DID I SAY THIS WAS THE END OF THE HACK... IT'S NOT, ALAS
	// (20180307/thisisaaronland)

	var lat = 37.7749;
	var lon = -122.4194;
	var zoom = 12;

	var hash = location.hash;
	hash = hash.substring(1,)	// remove leading #

	if (hash){

		var params = hash.split("&");
		var count = params.length;

		for (var i=0; i < count; i++){

		    var kv = params[i].split("=");

		    if (kv.length != 2){
			console.log("invalid parameter", kv)
			continue
		    }

		    if (kv[0] == "lat"){
			    lat = parseFloat(kv[1]);
		    }

		    else if (kv[0] == "lng"){
			    lon = parseFloat(kv[1]);
		    }

		    else if (kv[0] == "z"){
			zoom = parseInt(kv[1])
		    }

		    else {
			console.log("unsupported parameter", kv)
		    }
		}

	}

	map.setView([lat, lon], zoom);

	// I've never understood why I need to setView before invoking the hash thingy
	// in mapzen.js but it's a thing... (20180307/thisisaaronland)

	L.Mapzen.hash({
		map: map
	});

	slippymap.crosshairs.init(map);
	
	map.on("dragend", pip);
	pip();

	var jump_form = document.getElementById("jump-to-form");
	
	jump_form.onsubmit = function(){

		var input = document.getElementById("jump-to-latlon");

		jump_to(input.value);
		return false;
	};
});
