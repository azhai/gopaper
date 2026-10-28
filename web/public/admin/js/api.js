(function () {
  var API_BASE = '/api/admin';

  var auth = {
    getToken: function () { return localStorage.getItem('admin_token'); },
    setToken: function (token) { localStorage.setItem('admin_token', token); },
    logout: function () { localStorage.removeItem('admin_token'); },
    isLoggedIn: function () { return !!localStorage.getItem('admin_token'); },
  };

  function request(method, url, body) {
    var headers = {};
    if (body && !(body instanceof FormData)) headers['Content-Type'] = 'application/json';
    var token = auth.getToken();
    if (token) headers['Authorization'] = 'Bearer ' + token;

    var opts = { method: method, headers: headers };
    if (body) opts.body = body instanceof FormData ? body : JSON.stringify(body);

    return fetch(API_BASE + url, opts).then(function (response) {
      return response.json().catch(function () {
        return { code: -1, message: '服务器返回异常（' + response.status + '）' };
      }).then(function (result) {
        if (response.status === 401 && url !== '/login') {
          auth.logout();
          if (window.adminForceLogin) window.adminForceLogin();
          throw new Error(result.message || '登录已过期，请重新登录');
        }
        return result;
      });
    }).catch(function (err) {
      if (err instanceof TypeError) return { code: -1, message: '无法连接服务器，请检查网络' };
      throw err;
    });
  }

  window.adminAuth = auth;
  window.adminApi = {
    login: function (username, password) {
      return request('POST', '/login', { username: username, password: password });
    },
    getContents: function (page, pageSize, dir, type) {
      var url = '/articles?page=' + (page || 1) + '&pageSize=' + (pageSize || 100);
      if (dir) url += '&dir=' + encodeURIComponent(dir);
      if (type) url += '&type=' + type;
      return request('GET', url);
    },
    getContent: function (slug) {
      return request('GET', '/articles/' + encodeURIComponent(slug));
    },
    createContent: function (data) { return request('POST', '/articles', data); },
    updateContent: function (slug, data) {
      return request('PUT', '/articles/' + encodeURIComponent(slug), data);
    },
    deleteContent: function (slug) {
      return request('DELETE', '/articles/' + encodeURIComponent(slug));
    },
    previewMarkdown: function (content) {
      return request('POST', '/articles/preview', { content: content });
    },
    getImages: function (page, pageSize) {
      return request('GET', '/images?page=' + (page || 1) + '&pageSize=' + (pageSize || 24));
    },
    uploadImage: function (file) {
      var formData = new FormData();
      formData.append('file', file);
      return request('POST', '/images', formData);
    },
    deleteImage: function (fileName) {
      return request('DELETE', '/images/' + encodeURIComponent(fileName));
    },
    refreshCache: function () { return request('POST', '/cache/refresh'); },
    getDirs: function () { return request('GET', '/dirs'); },
    getSite: function () { return request('GET', '/settings'); },
    saveSite: function (data) { return request('PUT', '/settings', data); },
    getAnalytics: function (range) {
      return request('GET', '/analytics?range=' + encodeURIComponent(range || '1d'));
    },
    createSection: function (data) { return request('POST', '/sections', data); },
    updateSection: function (data) { return request('PUT', '/sections', data); },
    getThemes: function () { return request('GET', '/themes'); },
    previewTheme: function (data) { return request('POST', '/themes/preview', data); },
    createTheme: function (data) { return request('POST', '/themes', data); },
    updateTheme: function (id, data) {
      return request('PUT', '/themes/' + encodeURIComponent(id), data);
    },
    deleteTheme: function (id) {
      return request('DELETE', '/themes/' + encodeURIComponent(id));
    },
    activateTheme: function (id, enabledAt) {
      return request('POST', '/themes/' + encodeURIComponent(id) + '/activate', { enabledAt: enabledAt });
    },
  };
})();
