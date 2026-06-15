(function () {
  var prefetchCache = new Map();
  var prefetchLimit = 24;

  function startLoader() {
    var el = document.getElementById('page-loader');
    if (!el) return;
    el.classList.remove('opacity-0');
    var bar = el.querySelector('.page-loader-bar');
    if (bar) {
      bar.style.transform = 'scaleX(0.35)';
      bar.style.transition = 'transform 4s ease-out';
    }
  }

  function finishLoader() {
    var el = document.getElementById('page-loader');
    if (!el) return;
    var bar = el.querySelector('.page-loader-bar');
    if (bar) {
      bar.style.transition = 'transform 0.2s ease-out';
      bar.style.transform = 'scaleX(1)';
    }
    setTimeout(function () {
      el.classList.add('opacity-0');
      if (bar) {
        bar.style.transform = 'scaleX(0)';
        bar.style.transition = 'none';
      }
    }, 220);
  }

  function prefetchURL(url) {
    if (!url || prefetchCache.has(url) || url.indexOf('/lang/') !== -1) return;
    if (prefetchCache.size >= prefetchLimit) return;
    prefetchCache.set(url, true);
    var link = document.createElement('link');
    link.rel = 'prefetch';
    link.href = url;
    link.as = 'document';
    document.head.appendChild(link);
  }

  function bindPrefetchLinks(root) {
    (root || document).querySelectorAll('a.prefetch-link, a[href^="/"]:not([data-prefetch="off"])').forEach(function (a) {
      if (a.dataset.prefetchBound) return;
      if (!a.href || a.target === '_blank' || a.hasAttribute('download')) return;
      a.dataset.prefetchBound = '1';
      a.addEventListener('mouseenter', function () { prefetchURL(a.href); }, { passive: true });
      a.addEventListener('touchstart', function () { prefetchURL(a.href); }, { passive: true });
      a.addEventListener('click', startLoader, { passive: true });
    });
  }

  document.addEventListener('DOMContentLoaded', function () {
    var toggle = document.getElementById('sidebar-toggle');
    var sidebar = document.getElementById('sidebar');
    var overlay = document.getElementById('sidebar-overlay');
    if (toggle && sidebar) {
      function openSidebar() {
        sidebar.classList.remove('-translate-x-full');
        if (overlay) overlay.classList.remove('hidden');
        document.body.classList.add('overflow-hidden', 'lg:overflow-auto');
      }
      function closeSidebar() {
        sidebar.classList.add('-translate-x-full');
        if (overlay) overlay.classList.add('hidden');
        document.body.classList.remove('overflow-hidden', 'lg:overflow-auto');
      }
      toggle.addEventListener('click', function () {
        if (sidebar.classList.contains('-translate-x-full')) openSidebar();
        else closeSidebar();
      });
      if (overlay) overlay.addEventListener('click', closeSidebar);
      sidebar.querySelectorAll('a').forEach(function (a) {
        a.addEventListener('click', function () {
          if (window.innerWidth < 1024) closeSidebar();
        });
      });
    }

    bindPrefetchLinks(document);
    window.addEventListener('pageshow', finishLoader);
    finishLoader();
  });

  document.body.addEventListener('htmx:configRequest', function (evt) {
    var meta = document.querySelector('meta[name="csrf-token"]');
    if (meta && meta.content) {
      evt.detail.headers['X-CSRF-Token'] = meta.content;
    }
    startLoader();
  });

  document.body.addEventListener('htmx:afterSwap', function (evt) {
    finishLoader();
    bindPrefetchLinks(evt.detail.elt || document);
    var meta = document.querySelector('meta[name="csrf-token"]');
    if (!meta) return;
    document.querySelectorAll('input[name="csrf_token"]').forEach(function (input) {
      input.value = meta.content;
    });
  });

  document.body.addEventListener('htmx:responseError', finishLoader);
})();
