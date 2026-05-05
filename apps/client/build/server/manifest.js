const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set(["robots.txt"]),
	mimeTypes: {".txt":"text/plain"},
	_: {
		client: {start:"_app/immutable/entry/start.BInn8r2J.js",app:"_app/immutable/entry/app.sDe37yFV.js",imports:["_app/immutable/entry/start.BInn8r2J.js","_app/immutable/chunks/Dfc69r8K.js","_app/immutable/chunks/CqGZmWE0.js","_app/immutable/chunks/BX8VRQXo.js","_app/immutable/chunks/By0AxR0f.js","_app/immutable/entry/app.sDe37yFV.js","_app/immutable/chunks/CqGZmWE0.js","_app/immutable/chunks/JvAJrAwC.js","_app/immutable/chunks/CEYGRMaA.js","_app/immutable/chunks/By0AxR0f.js","_app/immutable/chunks/5EsqM6Ai.js","_app/immutable/chunks/CTaSuNox.js","_app/immutable/chunks/BX8VRQXo.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./chunks/0-DneNxdTM.js')),
			__memo(() => import('./chunks/1-BmYri2WJ.js')),
			__memo(() => import('./chunks/2-DKh9bcTG.js')),
			__memo(() => import('./chunks/3-BOGi2MZO.js')),
			__memo(() => import('./chunks/4-C5LFshZw.js'))
		],
		remotes: {
			
		},
		routes: [
			{
				id: "/",
				pattern: /^\/$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 2 },
				endpoint: null
			},
			{
				id: "/dashboard",
				pattern: /^\/dashboard\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 3 },
				endpoint: null
			},
			{
				id: "/login",
				pattern: /^\/login\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 4 },
				endpoint: null
			}
		],
		prerendered_routes: new Set([]),
		matchers: async () => {
			
			return {  };
		},
		server_assets: {}
	}
}
})();

const prerendered = new Set([]);

const base = "";

export { base, manifest, prerendered };
//# sourceMappingURL=manifest.js.map
