document.addEventListener('DOMContentLoaded', () => {
    // Контейнер для постов
    const feedContainer = document.getElementById('feedContainer');
    const feedLoader = document.getElementById('feedLoader');
    const allPostsBtn = document.getElementById('allPostsBtn');
    const logo = document.querySelector('.logo');

    // Текущий режим: null = общая лента, иначе ID пользователя
    let currentUserFeed = null;

    // Функция для получения заголовков авторизации
    function getAuthHeaders() {
        const token = localStorage.getItem('access_token');
        return token ? { 'Authorization': 'Bearer ' + token } : {};
    }

    // Загрузка постов с API
    async function fetchPosts(url) {
        feedLoader.style.display = 'block';
        feedContainer.innerHTML = '';
        try {
            const response = await fetch(url, { headers: getAuthHeaders() });
            if (!response.ok) throw new Error('Ошибка загрузки');
            const data = await response.json();
            if (data.success) {
                renderPosts(data.data);
            } else {
                feedContainer.innerHTML = `<p class="error-message">` +
                `{data.err_message || 'Не удалось загрузить посты'}</p>`;
            }
        } catch (err) {
            feedContainer.innerHTML = `<p class="error-message">${err.message}</p>`;
        } finally {
            feedLoader.style.display = 'none';
        }
    }

    // Рендер массива постов
    function renderPosts(posts) {
        if (!posts || posts.length === 0) {
            feedContainer.innerHTML = '<p style="text-align:center;padding:2rem;">'
            + 'Пока нет постов</p>';
            return;
        }

        feedContainer.innerHTML = posts.map(post => {
            // todo заглушка
            const imageUrl = post.image_url || 'https://via.placeholder.com/300x200?text=No+Image';
            return `
                <div class="post-card">
                    <img src="${imageUrl}" alt="Post image" class="post-image">
                    <div class="post-body">
                        <h3 class="post-title">${escapeHtml(post.title) || 'Без названия'}</h3>
                        <p class="post-description">${escapeHtml(post.description) || ''}</p>
                        <div class="post-meta">
                            <span class="post-author" data-user-id="${post.user_id}">${escapeHtml(post.username) || 'Аноним'}</span>
                            <span class="post-date">${post.created_at}</span>
                        </div>
                    </div>
                </div>
            `;
        }).join('');

        // Навешиваем обработчики клика по автору
        document.querySelectorAll('.post-author').forEach(el => {
            el.addEventListener('click', (e) => {
                const userId = e.target.dataset.userId;
                if (userId) {
                    loadUserFeed(parseInt(userId));
                }
            });
        });
    }

    // Простейшее экранирование HTML
    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Загрузка главной ленты
    function loadMainFeed() {
        currentUserFeed = null;
        allPostsBtn.style.display = 'none';
        fetchPosts('/loadMainFeed');
    }

    // Загрузка ленты конкретного пользователя
    function loadUserFeed(userId) {
        currentUserFeed = userId;
        allPostsBtn.style.display = 'inline-block';
        fetchPosts(`/loadUserFeed/${userId}`);
    }

    // Обработчики
    logo.addEventListener('click', (e) => {
        e.preventDefault();
        loadMainFeed();
    });

    allPostsBtn.addEventListener('click', () => {
        loadMainFeed();
    });

    // Клик по нику в хедере
    const userNameDisplay = document.getElementById('userNameDisplay');
    if (userNameDisplay) {
        userNameDisplay.addEventListener('click', () => {
            const user = JSON.parse(localStorage.getItem('user') || '{}');
            if (user && user.id) {
                loadUserFeed(user.id);
            }
        });
    }

    // Инициализация: загружаем главную ленту
    loadMainFeed();

    // Экспортируем функцию для вызова из других скриптов
    window.reloadFeed = function() {
        if (currentUserFeed === null) {
            loadMainFeed();
        } else {
            loadUserFeed(currentUserFeed);
        }
    };
});