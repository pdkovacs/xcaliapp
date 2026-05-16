export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    url.hostname = env.FUNCTION_URL_HOST;
    return fetch(url, request);
  },
};
