"use strict";
/* global __dirname */
/* global Buffer */
/*******************************************************************************
 * Copyright (c) 2016 IBM Corp.
 *
 * All rights reserved. 
 *   
 *******************************************************************************/
/*
	Version: 0.2
	Updated: 01/20/2016
*/

//Load modules
var fs = require('fs');
var path = require('path');
var https = require('https');
var http = require('http');
var async = require('async');
var rest = require(__dirname + "/lib/rest");
var AdmZip = require('adm-zip');
var chaincode_dir = path.join(__dirname,'../../chaincode');

var chaincode = {
					read: null,
					query: null,
					write: null,
					remove: null,
					deploy: null,
					details:{
								deployed_name: '',
								func: [],
								git_url: '',
								peers: [],
								vars: [],
								unzip_dir: '',
								zip_url: ''
					}
				};

function ibc() {}
ibc.selectedPeer = 0;
ibc.q = [];
ibc.lastPoll = 0;
ibc.lastBlock = 0;
var tempDirectory = path.join(__dirname, "./temp");									//	=./temp - temp directory name


// ============================================================================================================================
// EXTERNAL - load() - wrapper on a standard startup flow.
// 1. load network peer data
// 2. register users with security (if present)
// 3. load chaincode and parse
// ============================================================================================================================
ibc.prototype.load = function(options, cb){
	var errors = [];
	if(!options.network || !options.network.peers) errors.push("the option 'network.peers' is required");

	if(!options.chaincode || !options.chaincode.zip_url) errors.push("the option 'chaincode.zip_url' is required");
	if(!options.chaincode || !options.chaincode.unzip_dir) errors.push("the option 'chaincode.unzip_dir' is required");
	if(!options.chaincode || !options.chaincode.git_url) errors.push("the option 'chaincode.git_url' is required");
	if(errors.length > 0){															//check for input errors
		console.log('! [ibc-js] Input Error - ibc.load()', errors);
		if(cb) cb(eFmt('input error', 400, errors));
		return;																		//get out of dodge
	}
	
	// Step 1
	ibc.prototype.network(options.network.peers);
	
	// Step 2 - optional - only for secure networks
	console.log("[ibc-js] Commence registering users: #", options.network.users.length)
	if(options.network.users){
		//options.network.users = filter_users(options.network.users);				//only use the appropriate IDs

		var arr = [];
		for(var i in chaincode.details.peers){
			arr.push(i);															//build the list of indexes
		}
		async.each(arr, function(i, a_cb) {
			if(options.network.users[i]){											//make sure we still have a user for this network
				console.log("[ibc-js] Registering user: ", options.network.users[i].username);
				options.network.users.forEach(function(user, idx, arr){
				    ibc.prototype.register(i, user.username, user.secret);

				})
				a_cb();
			}
			else a_cb();
		}, function(err, data){
			load_cc();
		});
	}
	else{
		load_cc();
	}
	
	// Step 3
	function load_cc(){
		ibc.prototype.load_chaincode(options.chaincode, cb);						//download/parse and load chaincode
	}
};

// ============================================================================================================================
// EXTERNAL - load_chaincode() - load the chaincode and parssssssssse
// 0. Load the github or zip
// 1. Unzip & scan directory for files
// 2. Iter over go files
// 		2a. Find Run() in file, grab variable for *simplechaincode
//		2b. Grab function names that need to be exported
//		2c. Create JS function for golang function
// 3. Call callback()
// ============================================================================================================================
ibc.prototype.load_chaincode = function(options, cb) {
	var errors = [];
	if(!options.zip_url) errors.push("the option 'zip_url' is required");
	if(!options.unzip_dir) errors.push("the option 'unzip_dir' is required");
	if(!options.git_url) errors.push("the option 'git_url' is required");
	if(errors.length > 0){																//check for input errors
		console.log('! [ibc-js] Input Error - ibc.load_chaincode()', errors);
		if(cb) cb(eFmt('input error', 400, errors));
		return;																			//get out of dodge
	}
	
	var keep_looking = true;
	var zip_dest = path.join(tempDirectory,  '/file.zip');								//	=./temp/file.zip
	var unzip_dest = path.join(tempDirectory,  '/unzip');								//	=./temp/unzip
	var unzip_cc_dest = path.join(unzip_dest, '/', options.unzip_dir);					//	=./temp/unzip/DIRECTORY
	chaincode.details.zip_url = options.zip_url;
	chaincode.details.unzip_dir = options.unzip_dir;
	chaincode.details.git_url = options.git_url;

	console.log("zip_dest: "+zip_dest);
	console.log("unzip_dest: "+unzip_dest);
	console.log("unzip_cc_dest: "+unzip_cc_dest);
	console.log("chaincode.details.zip_url: "+chaincode.details.zip_url);
	console.log("chaincode.details.unzip_dir: "+chaincode.details.unzip_dir);
	console.log("chaincode.details.git_url: "+chaincode.details.git_url);
	console.log("chaincode_dir: "+chaincode_dir);

	if(!options.deployed_name || options.deployed_name == ''){							//lets clear and re-download
		ibc.prototype.clear(cb_ready);
	}
	else{
		chaincode.details.deployed_name = options.deployed_name;
		cb_ready();
	}
	
	// check if we already have the chaincode in the local filesystem, else download it
	function cb_ready(){
		// try{fs.mkdirSync(tempDirectory);}
		// catch(e){ }
		fs.access(chaincode_dir, cb_file_exists);										//check if files exist yet
		function cb_file_exists(e){
			if(e != null){
				download_it(options.zip_url);											//nope, go download it
			}
			else{
				console.log('[ibc-js] Found chaincode in local file system');
				fs.readdir(chaincode_dir, cb_got_names);								//yeppers, go use it
			}
		}
	}

	// Step 0.
	function download_it(download_url){
		console.log('[ibc-js] Downloading zip');
		var file = fs.createWriteStream(zip_dest);
		var handleResponse = function(response) {
			response.pipe(file);
			file.on('finish', function() {
				if(response.headers.status === '302 Found'){
					console.log('redirect...', response.headers.location);
					file.close();
					download_it(response.headers.location);
				}
				else{
					file.close(cb_downloaded);  									//close() is async
				}
			});
		}
		var handleError = function(err) {
			console.log('! [ibc-js] Download error');
			fs.unlink(zip_dest); 													//delete the file async
			if (cb) cb(eFmt('fs error', 500, err.message), chaincode);
		};

		var protocol = download_url.split('://')[0];
		if(protocol === 'https') {
			https.get(download_url, handleResponse).on('error', handleError);
		}
		else{
			http.get(download_url, handleResponse).on('error', handleError);
		}
	}
	
	// Step 1.
	function cb_downloaded(){
		console.log('[ibc-js] Unzipping zip');
		var zip = new AdmZip(zip_dest);
		zip.extractAllTo(unzip_dest, /*overwrite*/true);
		console.log('[ibc-js] Unzip done');
		fs.readdir(unzip_cc_dest, cb_got_names);
		fs.unlink(zip_dest, function(err) {});										//remove zip file, never used again
	}

	// Step 2.
	function cb_got_names(err, obj){
		console.log('[ibc-js] Scanning files', obj);
		var foundGo = false;
		if(err != null) console.log('! [ibc-js] fs readdir Error', err);
		else{
			for(var i in obj){
				if(obj[i].indexOf('.go') >= 0){										//look for GoLang files
					if(keep_looking){
						foundGo = true;
						fs.readFile(path.join(chaincode_dir, obj[i]), 'utf8', cb_read_go_file);
					}
				}
			}
		}
		if(!foundGo){																//error
			var msg = 'did not find any *.go files, cannot continue';
			console.log('! [ibc-js] Error - ', msg);
			if(cb) cb(eFmt('no chaincode', 400, msg), null);
		}
	}
	
	function cb_read_go_file(err, str){
		if(err != null) console.log('! [ibc-js] fs readfile Error', err);
		else{
			
			// Step 2a.
			var regex = /func\s+\((\w+)\s+\*SimpleChaincode\)\s+Run/i;				//find the variable name that Run is using for simplechaincode pointer
			var res = str.match(regex);
			if(!res || !res[1]){
				var msg = 'did not find Run() function in chaincode, cannot continue';
				console.log('! [ibc-js] Error -', msg);
				if(cb) cb(eFmt('missing run', 400, msg), null);
			}
			else{
				keep_looking = false;
				
				// Step 2b.
				var re = new RegExp('\\s' + res[1] + '\\.(\\w+)\\(', "gi");			//find the function names in Run()
				res = str.match(re);
				if(res[1] == null){
					console.log('[ibc-js] error did not find function names in chaincode');
				}
				else{
					
					// Step 2c.
					for(var i in res){												//build the rest call for each function
						var pos = res[i].indexOf('.');
						var temp = res[i].substring(pos + 1, res[i].length - 1);
						populate_go_chaincode(temp);
					}
					
					// Step 3.
					chaincode.read = read;
					chaincode.query = query;
					chaincode.write = write;
					chaincode.remove = remove;
					chaincode.deploy = deploy;
					if(cb) cb(null, chaincode);										//all done, send it to callback
				}
			}
		}
	}
};

// ============================================================================================================================
// EXTERNAL - network() - setup network configuration to hit a rest peer
// ============================================================================================================================
ibc.prototype.network = function(arrayPeers){
	var errors = [];
	if(!arrayPeers) errors.push("network input arg should be array of peer objects");
	else if(arrayPeers.constructor !== Array) errors.push("network input arg should be array of peer objects");
	if(errors.length > 0){															//check for input errors
		console.log('! [ibc-js] Input Error - ibc.network()', errors);
	}
	else{
		for(var i in arrayPeers){
			var pos = arrayPeers[i].id.indexOf('_') + 1;
			var temp = 	{
							name: '',
							api_host: arrayPeers[i].api_host,
							api_port: arrayPeers[i].api_port,
							id: arrayPeers[i].id,
							ssl: true
						};
			temp.name = arrayPeers[i].id.substring(pos) + '-' + arrayPeers[i].api_host + ':' + arrayPeers[i].api_port;	//build friendly name
			if(arrayPeers[i].api_url.indexOf('https') == -1) temp.ssl = false;
			console.log('[ibc-js] Peer: ', temp.name);
			chaincode.details.peers.push(temp);
		}

		rest.init({																	//load default values for rest call to peer
					host: chaincode.details.peers[0].api_host,
					port: chaincode.details.peers[0].api_port,
					headers: {
								"Content-Type": "application/json",
								"Accept": "application/json",
							},
					ssl: chaincode.details.peers[0].ssl,
					timeout: 60000,
					quiet: true
		});
	}
};


// ============================================================================================================================
// EXTERNAL - switchPeer() - switch the default peer to hit
// ============================================================================================================================
ibc.prototype.switchPeer = function(index) {
	if(chaincode.details.peers[index]) {
		rest.init({																	//load default values for rest call to peer
					host: chaincode.details.peers[index].api_host,
					port: chaincode.details.peers[index].api_port,
					headers: {
								"Content-Type": "application/json",
								"Accept": "application/json",
							},
					ssl: chaincode.details.peers[index].ssl,
					timeout: 60000,
					quiet: true
		});
		ibc.selectedPeer = index;
		return true;
	} else {
		return false;
	}
};

// ============================================================================================================================
// EXTERNAL - save() - write chaincode details to a json file
// ============================================================================================================================
ibc.prototype.save =  function(dir, cb){
	var errors = [];
	if(!dir) errors.push("the option 'dir' is required");
	if(errors.length > 0){																//check for input errors
		console.log('[ibc-js] Input Error - ibc.save()', errors);
		if(cb) cb(eFmt('input error', 400, errors));
	}
	else{
		var fn = 'chaincode.json';														//default name
		if(chaincode.details.deployed_name) fn = chaincode.details.deployed_name + '.json';
		var dest = path.join(dir, fn);
		fs.writeFile(dest, JSON.stringify({details: chaincode.details}), function(e){
			if(e != null){
				console.log(e);
				if(cb) cb(eFmt('fs write error', 500, e), null);
			}
			else {
				//console.log(' - saved ', dest);
				if(cb) cb(null, null);
			}
		});
	}
};

// ============================================================================================================================
// EXTERNAL - clear() - clear the temp directory
// ============================================================================================================================
ibc.prototype.clear =  function(cb){
	console.log('[ibc-js] removing temp dir');
	removeThing(tempDirectory, cb);
};

function removeThing(dir, cb){
	//console.log('!', dir);
	fs.readdir(dir, function (err, files) {
		if(err != null || !files || files.length == 0){
			cb();
		}
		else{
			async.each(files, function (file, cb) {						//over each thing
				file = path.join(dir, file);
				fs.stat(file, function(err, stat) {
					if (err) {
						if(cb) cb(err);
						return;
					}
					if (stat.isDirectory()) {
						removeThing(file, cb);							//keep going
					}
					else {
						//console.log('!', dir);
						fs.unlink(file, function(err) {
							if (err) {
								//console.log('error', err);
								if(cb) cb(err);
								return;
							}
							//console.log('good', dir);
							if(cb) cb();
							return;
						});
					}
				});
			}, function (err) {
				if(err){
					if(cb) cb(err);
					return;
				}
				fs.rmdir(dir, function (err) {
					if(cb) cb(err);
					return;
				});
			});
		}
	});
}

//============================================================================================================================
// EXTERNAL chain_stats() - get blockchain stats
//============================================================================================================================
ibc.prototype.chain_stats =  function(cb){
	var options = {path: '/chain'};

	options.success = function(statusCode, data){
		console.log("[ibc-js] Chain Stats - success");
		if(cb) cb(null, data);
	};
	options.failure = function(statusCode, e){
		console.log("[ibc-js] Chain Stats - failure:", statusCode);
		if(cb) cb(eFmt('http error', statusCode, e), null);
	};
	rest.get(options, '');
};

//============================================================================================================================
// EXTERNAL block_stats() - get block meta data
//============================================================================================================================
ibc.prototype.block_stats =  function(id, cb){
	var options = {path: '/chain/blocks/' + id};					//i think block IDs start at 0, height starts at 1, fyi
	options.success = function(statusCode, data){
		console.log("[ibc-js] Block Stats - success");
		if(cb) cb(null, data);
	};
	options.failure = function(statusCode, e){
		console.log("[ibc-js] Block Stats - failure:", statusCode);
		if(cb) cb(eFmt('http error', statusCode, e), null);
	};
	rest.get(options, '');
};


//============================================================================================================================
//read() - read generic variable from chaincode state
//============================================================================================================================
function read(name, cb, lvl){										//lvl is for reading past state blocks, tbd exactly
	var options = {
		path: '/devops/query'
	};
	var body = {
					chaincodeSpec: {
						type: "GOLANG",
						chaincodeID: {
							name: chaincode.details.deployed_name,
						},
						ctorMsg: {
							function: "query",
							args: [name]
						},
						secureContext: chaincode.details.peers[ibc.selectedPeer].user
					}
				};
	console.log('body', body);
	options.success = function(statusCode, data){
		console.log("[ibc-js] Read - success:", data);
		if(cb) cb(null, data.OK);
	};
	options.failure = function(statusCode, e){
		console.log("[ibc-js] Read - failure:", statusCode);
		if(cb) cb(eFmt('http error', statusCode, e), null);
	};
	rest.post(options, '', body);
}

//============================================================================================================================
//query() - read generic variable from chaincode state
//============================================================================================================================
function query(args, cb, lvl){										//lvl is for reading past state blocks, tbd exactly
	var options = {
		path: '/devops/query'
	};
	var body = {
					chaincodeSpec: {
						type: "GOLANG",
						chaincodeID: {
							name: chaincode.details.deployed_name,
						},
						ctorMsg: {
							function: "query",
							args: args
						},
						secureContext: chaincode.details.peers[ibc.selectedPeer].user
					}
				};
	console.log('body', body);
	options.success = function(statusCode, data){
		console.log("[ibc-js] Query - success:", data);
		if(cb) cb(null, data.OK);
	};
	options.failure = function(statusCode, e){
		console.log("[ibc-js] Query - failure:", statusCode);
		if(cb) cb(eFmt('http error', statusCode, e), null);
	};
	rest.post(options, '', body);
}

//============================================================================================================================
//write() - write generic variable to chaincode state
//============================================================================================================================
function write(name, val, cb){
	var options = {
		path: '/devops/invoke'
	};
	var body = {
					chaincodeSpec: {
						type: "GOLANG",
						chaincodeID: {
							name: chaincode.details.deployed_name,
						},
						ctorMsg: {
							function: 'write',
							args: [name, val]
						},
						secureContext: chaincode.details.peers[ibc.selectedPeer].user
					}
				};
	
	options.success = function(statusCode, data){
		console.log("[ibc-js] Write - success:", data);
		ibc.q.push(Date.now());
		if(cb) cb(null, data);
	};
	options.failure = function(statusCode, e){
		console.log("[ibc-js] Write - failure:", statusCode);
		if(cb) cb(eFmt('http error', statusCode, e), null);
	};
	rest.post(options, '', body);
}

//============================================================================================================================
//remove() - delete a generic variable from chaincode state
//============================================================================================================================
function remove(name, cb){
	var options = {
		path: '/devops/invoke'
	};
	var body = {
					chaincodeSpec: {
						type: "GOLANG",
						chaincodeID: {
							name: chaincode.details.deployed_name,
						},
						ctorMsg: {
							function: 'delete',
							args: [name]
						},
						secureContext: chaincode.details.peers[ibc.selectedPeer].user
					}
				};

	options.success = function(statusCode, data){
		console.log("[ibc-js] Remove - success:", data);
		ibc.q.push(Date.now());
		if(cb) cb(null, data);
	};
	options.failure = function(statusCode, e){
		console.log("[ibc-js] Remove - failure:", statusCode);
		if(cb) cb(eFmt('http error', statusCode, e), null);
	};
	rest.post(options, '', body);
}

//============================================================================================================================
//register() - register a username with a peer (only for a secured blockchain network)
//============================================================================================================================
ibc.prototype.register = function(index, enrollID, enrollSecret, cb) {
	console.log("[ibc-js] Registering ", chaincode.details.peers[index].name, " w/enrollID - " + enrollID);
	var options = {
		path: '/registrar',
		host: chaincode.details.peers[index].api_host,
		port: chaincode.details.peers[index].api_port,
		ssl: chaincode.details.peers[index].ssl
	};

	var body = 	{
					enrollId: enrollID,
					enrollSecret: enrollSecret
				};

	options.success = function(statusCode, data){
		console.log("[ibc-js] Registration success:", enrollID);
		chaincode.details.peers[index].user = enrollID;							//remember the user for this peer
		if(cb){
			cb(null, data);
		}
	};
	options.failure = function(statusCode, e){
		console.log("[ibc-js] Register - failure:", enrollID, statusCode);
		if(cb) cb(eFmt('http error', statusCode, e), null);
	};
	rest.post(options, '', body);
};

//============================================================================================================================
//deploy() - deploy chaincode and call a cc function
//============================================================================================================================
function deploy(func, args, save_path, cb){
	console.log("[ibc-js] Deploying Chaincode - Start");
	console.log("\n\n\t Waiting...");										//this can take awhile
	var options = {path: '/devops/deploy', timeout: 80000};
	var body = 	{
					type: "GOLANG",
					chaincodeID: {
							path: chaincode.details.git_url
						},
					ctorMsg:{
							"function": func,
							"args": args
					},
					secureContext: chaincode.details.peers[ibc.selectedPeer].user
				};
	console.log('!body', body);
	options.success = function(statusCode, data){
		console.log("\n\n\t deploy success [wait 1 more minute]");
		chaincode.details.deployed_name = data.message;
		ibc.prototype.save(tempDirectory);									//save it so we remember we have deployed
		if(save_path != null) ibc.prototype.save(save_path);				//user wants the updated file somewhere
		if(cb){
			setTimeout(function(){
				console.log("[ibc-js] Deploying Chaincode - Complete");
				cb(null, data);
			}, 40000);														//wait extra long, not always ready yet
		}
	};
	options.failure = function(statusCode, e){
		console.log("[ibc-js] deploy - failure:", statusCode);
		if(cb) cb(eFmt('http error', statusCode, e), null);
	};
	rest.post(options, '', body);
}

//============================================================================================================================
//heart_beat() - interval function to poll against blockchain height (has fast and slow mode)
//============================================================================================================================
var slow_mode = 10000;
var fast_mode = 500;
function heart_beat(){
	if(ibc.lastPoll + slow_mode < Date.now()){								//slow mode poll
		//console.log('[ibc-js] Its been awhile, time to poll');
		ibc.lastPoll = Date.now();
		ibc.prototype.chain_stats(cb_got_stats);
	}
	else{
		for(var i in ibc.q){
			var elasped = Date.now() - ibc.q[i];
			if(elasped <= 3000){											//fresh unresolved action, fast mode!
				console.log('[ibc-js] Unresolved action, must poll');
				ibc.lastPoll = Date.now();
				ibc.prototype.chain_stats(cb_got_stats);
			}
			else{
				//console.log('[ibc-js] Expired, removing');
				ibc.q.pop();												//expired action, remove it
			}
		}
	}
}

function cb_got_stats(e, stats){
	if(e == null){
		if(stats && stats.height){
			if(ibc.lastBlock != stats.height) {									//this is a new block!
				console.log('[ibc-js] New block!', stats.height);
				ibc.lastBlock  = stats.height;
				ibc.q.pop();													//action is resolved, remove
				if(ibc.monitorFunction) ibc.monitorFunction(stats);				//call the user's callback
			}
		}
	}
}

ibc.prototype.monitor_blockheight = function(cb) {							//hook in your own function, triggers when chain grows
	setInterval(function(){heart_beat();}, fast_mode);
	ibc.monitorFunction = cb;												//store it
};



//============================================================================================================================
//													Helper Functions() 
//============================================================================================================================
//populate_chaincode() - create JS call for custom goLang function, stored in chaincode var!
//==================================================================
function populate_go_chaincode(name){
	if(chaincode[name] != null){
		//console.log('[ibc-js] \t skip, already exists');					//skip
	}
	else {
		console.log('[ibc-js] Found cc function: ', name);
		chaincode.details.func.push(name);
		chaincode[name] = function(args, cb){								//create the functions in chaincode obj
			var options = {path: '/devops/invoke'};
			var body = {
					chaincodeSpec: {
						type: "GOLANG",
						chaincodeID: {
							name: chaincode.details.deployed_name,
						},
						ctorMsg: {
							function: name,
							args: args
						},
						secureContext: chaincode.details.peers[ibc.selectedPeer].user
					}
			};

			options.success = function(statusCode, data){
				console.log("[ibc-js]", name, " - success:", data);
				ibc.q.push(Date.now());
				if(cb) cb(null, data);
			};
			options.failure = function(statusCode, e){
				console.log("[ibc-js]", name, " - failure:", statusCode);
				if(cb) cb(eFmt('http error', statusCode, e), null);
			};
			rest.post(options, '', body);
		};
	}
}

//==================================================================
//filter_users() - only get client level usernames - [1=client, 2=nvp, 4=vp, 8=auditor accurate as of 2/18]
//==================================================================
function filter_users(users){

	var valid_users = [];
	for(var i = 0; i < users.length; i++) {

		if(users[i].usertype == 1 || users[i].usertype == 2 || users[i].usertype == 3){		//type should be 1 for client
			valid_users.push(users[i]);
		}
	}
	return valid_users;
}

//==================================================================
//eFmt() - format errors
//==================================================================
function eFmt(name, code, details){
	return 	{
		name: String(name),
		code: Number(code),
		details: details
	};
}



module.exports = ibc;