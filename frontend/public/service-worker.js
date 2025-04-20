const CACHE_NAME = 'pwa-cache-v1';
const FILES_TO_CACHE = [
    '/index.html',
    '/favicon.ico',
    '/manifest.webmanifest',
];


async function cacheFirst(request) {
    try {
        const cachedResponse = await caches.match(request);
        if (cachedResponse) {
            console.log('Serving from cache:', request.url);
            return cachedResponse;
        }
        console.log('Fetching from network:', request.url);
        const networkResponse = await fetch(request);
        if (networkResponse.ok) {
            const cache = await caches.open(CACHE_NAME);
            await cache.put(request, networkResponse.clone());
        }
        return networkResponse;
    } catch (error) {
        console.error('CacheFirst error:', error);
        return new Response('Offline', { status: 503, statusText: 'Service Unavailable' });
    }
}

async function networkFirst(request) {
    const cache = await caches.open(CACHE_NAME);

    try {
        const networkResponse = await fetch(request);
        if (networkResponse.ok) {
            await cache.put(request, networkResponse.clone());
        }
        return networkResponse;
    } catch (error) {
        console.warn('Network request failed, serving from cache:', request.url);
        const cachedResponse = await caches.match(request);
        return cachedResponse || new Response('Offline', { status: 503, statusText: 'Service Unavailable' });
    }
}


// eslint-disable-next-line no-restricted-globals
self.addEventListener('install', (event) => {
    console.log('Service Worker: Installing...');
    // eslint-disable-next-line no-restricted-globals
    self.skipWaiting();
    event.waitUntil(
        caches.open(CACHE_NAME).then((cache) => {
            console.log('Service Worker: Caching Files');
            return fetch('asset-manifest.json').then((response) => {
                return response.json();
            }).then((manifest) => {
                const filesToCache = [...FILES_TO_CACHE];

                Object.keys(manifest.files).forEach((filePath) => {
                    if (filePath.endsWith('.css') || filePath.endsWith('.js')) {
                        filesToCache.push(manifest.files[filePath]);
                    }
                });

                /*
                // API that must return json files for caching
                const apiUrls = [];
                const fetchPromises = apiUrls.map((url) => {
                    return fetch(url).then((response) => {
                        if (response.ok) {
                            return cache.put(url, response.clone());
                        }
                    });
                });
                */
                return Promise.all([
                    //...fetchPromises,
                    cache.addAll(filesToCache)
                ]);
            });
        })
    );
});


// eslint-disable-next-line no-restricted-globals
self.addEventListener('activate', (event) => {
    console.log('Service Worker: Activated');
    const cacheWhitelist = [CACHE_NAME];
    event.waitUntil(
        caches.keys().then((cacheNames) => {
            return Promise.all(
                cacheNames.map((cacheName) => {
                    if (!cacheWhitelist.includes(cacheName)) {
                        console.log(`Service Worker: Deleting cache ${cacheName}`);
                        return caches.delete(cacheName);
                    }
                    return undefined;
                })
            );
        })
    );
});

// eslint-disable-next-line no-restricted-globals
self.addEventListener('message', (event) => {
    if (event.data.action === 'triggerPush') {
        const messageBody = event.data.body || 'Default notification body';

        // eslint-disable-next-line no-restricted-globals
        self.registration.showNotification('Triggered Push Notification', {
            body: messageBody,
            icon: '/favicon.ico',
        });
    }
});

// eslint-disable-next-line no-restricted-globals
self.addEventListener('fetch', (event) => {
    if (event.request.mode === 'navigate') { // this is necessary to use the same HTML file (metatags aren't important)
        event.respondWith(
            caches.match('en/index.html').then(response => response || fetch(event.request))
        );
        return;
    }

    const url = new URL(event.request.url);
    if (url.pathname.startsWith('/static/json/')) { // Cache First
        event.respondWith(cacheFirst(event.request));
    } else if (url.pathname === '/index.html' || url.pathname.match(/\.(css|js|ico)$/)) { // Cache First
        event.respondWith(cacheFirst(event.request));
    } else if (url.pathname.startsWith('/api/')) {
        event.respondWith(networkFirst(event.request)); // Network first later
    }
});

// eslint-disable-next-line no-restricted-globals
self.addEventListener('push', event => {
    console.log('Service Worker: Push Received.');

    let data = {};
    try {
        data = event.data ? event.data.json() : {};
    } catch (e) {
        console.warn('Push event but no JSON payload', e);
    }

    const title = data.title || 'nott';
    const options = {
        body: data.body ||  'events triggered',
        icon: data.icon || '/favicon.ico',
        badge: data.badge || '/favicon.ico',
        data: data.data || {},
        actions: data.actions || []
    };

    event.waitUntil(
        // eslint-disable-next-line no-restricted-globals
        self.registration.showNotification(title, options)
    );
});

// eslint-disable-next-line no-restricted-globals
self.addEventListener('notificationclick', event => { // обработка клика на уведомление(тяжко)
    console.log('Notification was clicked');
});
