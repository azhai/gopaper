(function () {
  'use strict';

  var STORAGE_KEY = 'gopaper_cart_v1';

  function safePath(value) {
    var path = String(value || '');
    if (path.charAt(0) === '/' && path.charAt(1) !== '/') return path;
    return '';
  }

  function safeUrl(value) {
    if (!value) return '';
    try {
      var url = new URL(String(value), window.location.origin);
      if (url.protocol === 'http:' || url.protocol === 'https:') return url.href;
    } catch (err) {
      return '';
    }
    return '';
  }

  function normalizeItem(raw) {
    var spec = String(raw.spec || '');
    var price = String(raw.price || '');
    var sku = String(raw.sku || '');
    return {
      key: String(raw.slug + '|' + (raw.sku || raw.spec || '')),
      slug: String(raw.slug),
      title: String(raw.title),
      href: safePath(raw.href),
      spec: spec,
      price: price,
      sku: sku,
      cover: safePath(raw.cover),
      jdUrl: safeUrl(raw.jdUrl),
      taobaoUrl: safeUrl(raw.taobaoUrl),
      meta: [spec, price, sku ? 'SKU ' + sku : ''].filter(Boolean).join(' · ') || '未选择规格'
    };
  }

  function isCartItem(item) {
    return item && typeof item === 'object' && item.slug && item.title;
  }

  function readCart() {
    try {
      var raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return [];
      var parsed = JSON.parse(raw);
      if (!Array.isArray(parsed)) return [];
      return parsed.filter(isCartItem).map(normalizeItem);
    } catch (err) {
      return [];
    }
  }

  document.addEventListener('alpine:init', function () {
    Alpine.store('cart', {
      items: readCart(),
      selected: {},
      open: false,
      status: '',

      get count() {
        return this.items.length;
      },

      get fabLabel() {
        return this.items.length > 0 ? '打开购物车，共 ' + this.items.length + ' 件产品' : '打开购物车';
      },

      isSelected(key) {
        return this.selected[key] !== false;
      },

      selectedItems() {
        var selected = this.selected;
        return this.items.filter(function (item) {
          return selected[item.key] !== false;
        });
      },

      platformCount(platform) {
        var field = platform === 'jd' ? 'jdUrl' : 'taobaoUrl';
        return this.selectedItems().filter(function (item) {
          return !!item[field];
        }).length;
      },

      add(base, spec) {
        var item = normalizeItem({
          slug: base.slug,
          title: base.title,
          href: base.href,
          cover: base.cover,
          spec: spec.name,
          price: spec.price,
          sku: spec.sku,
          jdUrl: spec.jdUrl,
          taobaoUrl: spec.taobaoUrl
        });
        var index = this.items.findIndex(function (entry) { return entry.key === item.key; });
        if (index >= 0) this.items[index] = item;
        else this.items.push(item);
        this.selected[item.key] = true;
        this.persist();
        return item;
      },

      remove(key) {
        this.items = this.items.filter(function (item) { return item.key !== key; });
        delete this.selected[key];
        this.persist();
      },

      toggle(key, checked) {
        this.selected[key] = !!checked;
      },

      setOpen(open) {
        this.open = !!open;
        document.body.style.overflow = this.open ? 'hidden' : '';
        if (this.open) this.status = '';
      },

      checkout(platform) {
        var items = this.selectedItems();
        var label = platform === 'jd' ? '京东' : '淘宝';
        var urls = items.map(function (item) {
          return platform === 'jd' ? item.jdUrl : item.taobaoUrl;
        }).filter(Boolean);

        if (urls.length === 0) {
          this.status = '所选产品没有' + label + '链接，请取消勾选或联系商家。';
          return;
        }

        var opened = 0;
        urls.forEach(function (url) {
          var win = null;
          try {
            win = window.open(url, '_blank');
          } catch (err) {
            win = null;
          }
          if (win) {
            try { win.opener = null; } catch (err) { /* ignore */ }
            opened++;
          }
        });

        var missing = items.length - urls.length;
        var message = '已尝试打开 ' + urls.length + ' 个' + label + '链接，成功 ' + opened + ' 个。';
        if (opened < urls.length) {
          message += ' 其余链接被浏览器拦截，请允许本站弹窗后重试，或复制清单手动打开。';
        }
        if (missing > 0) {
          message += ' 另有 ' + missing + ' 个所选产品没有' + label + '链接。';
        }
        this.status = message;
      },

      copyChecklist() {
        var items = this.selectedItems();
        if (items.length === 0) {
          this.status = '请先勾选要下单的产品。';
          return;
        }
        var self = this;
        var lines = items.map(function (item) {
          return [
            item.title,
            item.spec,
            item.price,
            item.sku ? 'SKU ' + item.sku : '',
            item.jdUrl || '',
            item.taobaoUrl || ''
          ].filter(Boolean).join(' | ');
        });
        var text = '选品清单\n' + lines.join('\n');

        function done() { self.status = '已复制 ' + items.length + ' 个产品的选品清单。'; }
        function failed() { self.status = '复制失败，请手动选择文字复制。'; }

        if (navigator.clipboard && window.isSecureContext) {
          navigator.clipboard.writeText(text).then(done, failed);
          return;
        }
        var area = document.createElement('textarea');
        area.value = text;
        area.setAttribute('readonly', '');
        area.style.position = 'fixed';
        area.style.opacity = '0';
        document.body.appendChild(area);
        area.select();
        var ok = false;
        try { ok = document.execCommand('copy'); } catch (err) { ok = false; }
        document.body.removeChild(area);
        if (ok) done(); else failed();
      },

      persist() {
        try {
          localStorage.setItem(STORAGE_KEY, JSON.stringify(this.items));
        } catch (err) {
          // Storage may be unavailable in private mode; the cart still works for this page.
        }
      }
    });

    Alpine.data('cartDrawer', function () {
      return {
        lastFocus: null,

        init() {
          this.$watch('$store.cart.open', this.onOpenChange.bind(this));
        },

        onOpenChange(open) {
          if (open) {
            this.lastFocus = document.activeElement;
            this.$nextTick(function () {
              var close = this.$el.querySelector('.cart-close');
              if (close) close.focus();
            }.bind(this));
            return;
          }
          if (this.lastFocus && typeof this.lastFocus.focus === 'function') {
            this.lastFocus.focus();
          }
          this.lastFocus = null;
        },

        close() {
          this.$store.cart.setOpen(false);
        },

        onKeydown(event) {
          if (!this.$store.cart.open) return;
          if (event.key === 'Escape') {
            event.preventDefault();
            this.close();
            return;
          }
          if (event.key !== 'Tab') return;

          var panel = this.$el.querySelector('.cart-panel');
          if (!panel) return;
          var focusable = Array.prototype.filter.call(
            panel.querySelectorAll('a[href], button:not([disabled]), input:not([disabled])'),
            function (el) { return el.offsetParent !== null; }
          );
          if (focusable.length === 0) return;
          var first = focusable[0];
          var last = focusable[focusable.length - 1];
          if (event.shiftKey && document.activeElement === first) {
            event.preventDefault();
            last.focus();
          } else if (!event.shiftKey && document.activeElement === last) {
            event.preventDefault();
            first.focus();
          }
        }
      };
    });

    Alpine.data('productBuy', function () {
      return {
        active: 0,
        specs: [],
        base: { slug: '', title: '', href: '', cover: '' },
        added: -1,
        note: '',
        timer: 0,

        init() {
          var el = this.$el;
          this.base = {
            slug: el.dataset.slug || '',
            title: el.dataset.title || '',
            href: el.dataset.href || '',
            cover: el.dataset.cover || ''
          };
          this.specs = Array.prototype.map.call(el.querySelectorAll('.spec-panel'), function (panel) {
            return {
              index: Number(panel.dataset.index),
              name: panel.dataset.name || '',
              price: panel.dataset.price || '',
              sku: panel.dataset.sku || '',
              jdUrl: panel.dataset.jd || '',
              taobaoUrl: panel.dataset.taobao || ''
            };
          });
          var checked = el.querySelector('input[name="product-spec"]:checked');
          this.active = checked ? Number(checked.value) : 0;
        },

        get current() {
          var active = this.active;
          return this.specs.find(function (spec) {
            return spec.index === active;
          }) || null;
        },

        get priceLabel() {
          return (this.current && this.current.price) || '询价';
        },

        addToCart(index) {
          var spec = this.specs.find(function (entry) { return entry.index === index; });
          if (!spec) return;
          var item = this.$store.cart.add(this.base, spec);
          this.note = '已加入购物车：' + item.title + (item.spec ? ' · ' + item.spec : '');
          this.added = index;
          window.clearTimeout(this.timer);
          this.timer = window.setTimeout(function () {
            this.added = -1;
          }.bind(this), 1500);
        },

        destroy() {
          window.clearTimeout(this.timer);
        }
      };
    });

  });
})();
