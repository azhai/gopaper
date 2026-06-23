import m from 'mithril';
const API_BASE = '/api/admin';
async function request(method, url, body) {
    const token = authService.getToken();
    const headers = {};
    if (body && !(body instanceof FormData)) {
        headers['Content-Type'] = 'application/json';
    }
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }
    const opts = { method, headers };
    if (body) {
        opts.body = body instanceof FormData ? body : JSON.stringify(body);
    }
    const response = await fetch(`${API_BASE}${url}`, opts);
    if (response.status === 401) {
        authService.logout();
        m.route.set('/login');
        throw new Error('会话已过期');
    }
    const result = await response.json();
    return result;
}
export const apiService = {
    login: (username, password) => request('POST', '/login', { username, password }),
    getArticles: (page = 1, pageSize = 10, dir, type) => {
        let url = `/articles?page=${page}&pageSize=${pageSize}`;
        if (dir)
            url += `&dir=${encodeURIComponent(dir)}`;
        if (type)
            url += `&type=${type}`;
        return request('GET', url);
    },
    getArticle: (slug) => request('GET', `/articles/${slug}`),
    createArticle: (data) => request('POST', '/articles', data),
    updateArticle: (slug, data) => request('PUT', `/articles/${slug}`, data),
    deleteArticle: (slug) => request('DELETE', `/articles/${slug}`),
    previewArticle: (content) => request('POST', '/articles/preview', { content }),
    getImages: (page = 1, pageSize = 20) => request('GET', `/images?page=${page}&pageSize=${pageSize}`),
    uploadImage: (file) => {
        const formData = new FormData();
        formData.append('file', file);
        return request('POST', '/images', formData);
    },
    deleteImage: (fileName) => request('DELETE', `/images/${fileName}`),
    refreshCache: () => request('POST', '/cache/refresh'),
    getDirs: () => request('GET', '/dirs'),
    getSettings: () => request('GET', '/settings'),
    updateSettings: (data) => request('PUT', '/settings', data),
    getLayouts: () => request('GET', '/layouts'),
    saveLayouts: (data) => request('PUT', '/layouts', data),
    getRegions: (name) => request('GET', `/layouts/${name}/regions`),
};
export const authService = {
    getToken: () => localStorage.getItem('admin_token'),
    setToken: (token) => localStorage.setItem('admin_token', token),
    logout: () => localStorage.removeItem('admin_token'),
    isLoggedIn: () => !!localStorage.getItem('admin_token'),
};
