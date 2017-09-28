window.addEventListener("load", function load(event){

	var map;
	var marker;			    
	var candidates;
	var intersecting = [];
	
	var draw_coords = function(){
		
		var center = map.getCenter();
		var lat = center.lat;
		var lon = center.lng;
		
		lat = lat.toFixed(6);
		lon = lon.toFixed(6);
		
		var coords = document.getElementById("coords");
		coords.innerText = lat + ", " + lon;
		
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

			console.log("INTERSECTING", rsp);
			
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
				
				var id = spr["wof:id"];
				var name = spr["wof:name"];
				
				var c = document.getElementById("candidate-" + id);

				if (c){
					c.appendChild(document.createTextNode(" " + name));
					c.setAttribute("class", "intersects");
				}
				
				fetch_geojson(url);
			}
       		};
		
		fetch(url, onsuccess);
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
	
	L.Mapzen.apiKey = document.body.getAttribute("data-mapzen-api-key");

	var map_opts = {
		tangramOptions: {
                        scene: "/tangram/refill-style.zip",
			tangramURL: "/javascript/tangram.js",
                }
	};
	
	map = L.Mapzen.map('map', map_opts);
	map.setView([37.7749, -122.4194], 12);
	
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
	
	map.on('move', draw_coords);
	map.on('dragend', pip);
	
	draw_coords();
	pip();

});
