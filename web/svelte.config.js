import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltekit/kit').Config} */
const config = {
	kit: {
		adapter: adapter({
			fallback: 'index.html',
			precompress: false,
			strict: true
		})
	}
};

export default config;
