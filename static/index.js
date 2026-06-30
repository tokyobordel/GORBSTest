document.addEventListener('DOMContentLoaded', () => {
    // Модальные окна
    const signinModal = document.getElementById('signinModal');
    const signupModal = document.getElementById('signupModal');
    const btnSignin = document.getElementById('btnSignin');
    const btnSignup = document.getElementById('btnSignup');
    const switchToSignup = document.getElementById('switchToSignup');
    const switchToSignin = document.getElementById('switchToSignin');
    const closeButtons = document.querySelectorAll('.close');

    // DOM-элементы для управления UI
    const guestButtons = document.getElementById('guestButtons');
    const userBlock = document.getElementById('userBlock');
    const userNameDisplay = document.getElementById('userNameDisplay');
    const btnLogout = document.getElementById('btnLogout');

    // Формы
    const signinForm = document.getElementById('signinForm');
    const signupForm = document.getElementById('signupForm');

    // Блоки ошибок
    const signinError = document.getElementById('signinError');
    const signupError = document.getElementById('signupError');

    // Сброс формы и сообщения об ошибке
    window.resetForm = function(form, errorElement) {
        form.reset();
        if (errorElement) errorElement.textContent = '';
    };

    // Открытие модального окна
    window.openModal = function(modal) {
        modal.classList.add('active');
    };

    // Закрытие модального окна со сбросом формы
    window.closeModal = function(modal) {
        modal.classList.remove('active');
        // сброс всех форм в модалке
        const forms = modal.querySelectorAll('form');
        forms.forEach(f => f.reset());
        // очистка ошибок
        modal.querySelectorAll('.error-message').forEach(el => el.textContent = '');
        // очистка списка файлов (если есть)
        const fileList = modal.querySelector('#fileList');
        if (fileList) fileList.innerHTML = '';
    };

    // Обработчики открытия
    btnSignin.addEventListener('click', () => openModal(signinModal));
    btnSignup.addEventListener('click', () => openModal(signupModal));

    // Переключение между окнами
    switchToSignup.addEventListener('click', (e) => {
        e.preventDefault();
        closeModal(signinModal);
        openModal(signupModal);
    });

    switchToSignin.addEventListener('click', (e) => {
        e.preventDefault();
        closeModal(signupModal);
        openModal(signinModal);
    });

    // Закрытие по крестику
    closeButtons.forEach(btn => {
        btn.addEventListener('click', (e) => {
            const modal = e.target.closest('.modal');
            closeModal(modal);
        });
    });

    // Закрытие по клику вне окна
    window.addEventListener('click', (e) => {
        if (e.target.classList.contains('modal')) {
            closeModal(e.target);
        }
    });

    // --- Отправка формы регистрации ---
    signupForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        signupError.textContent = '';

        const username = document.getElementById('signupUsername').value.trim();
        const email = document.getElementById('signupEmail').value.trim();
        const password = document.getElementById('signupPassword').value;
        const confirm = document.getElementById('signupConfirm').value;
        const tgChatId = document.getElementById('signupTg').value.trim();

        // Валидация
        if (!username || !email || !password || !confirm) {
            signupError.textContent = 'Заполните все обязательные поля';
            return;
        }

        if (password !== confirm) {
            signupError.textContent = 'Пароли не совпадают';
            return;
        }

        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(email)) {
            signupError.textContent = 'Введите корректный email';
            return;
        }

        const payload = {
            username,
            password,
            email,
            tg_chat_id: tgChatId || undefined
        };

        try {
            const response = await fetch('/signup', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });

            const data = await response.json();

            if (!response.ok || !data.success) {
                throw new Error(data.err_message || 'Ошибка регистрации');
            }

            closeModal(signupModal); // форма сбросится
        } catch (err) {
            signupError.textContent = err.message;
        }
    });

    // Отправка формы входа
    signinForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        signinError.textContent = '';

        const username = document.getElementById('signinUsername').value.trim();
        const password = document.getElementById('signinPassword').value;

        if (!username || !password) {
            signinError.textContent = 'Заполните все поля';
            return;
        }

        try {
            const response = await fetch('/signin', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password })
            });

            const data = await response.json();

            if (!response.ok || !data.success) {
                throw new Error(data.err_message || 'Ошибка входа');
            }

            // data.data должен содержать { access_token, refresh_token, user }
            const { access_token, refresh_token, user } = data.data;
            if (!access_token || !refresh_token || !user) {
                throw new Error('Некорректный ответ сервера');
            }

            // Сохраняем сессию
            saveSession(access_token, refresh_token, user);

            // Обновляем UI
            showLoggedInUI(user);

            closeModal(signinModal);
        } catch (err) {
            signinError.textContent = err.message;
        }
    });

    btnLogout.addEventListener('click', async () => {
        const refreshToken = localStorage.getItem('refresh_token');
        if (refreshToken) {
            try {
                await fetch('/logout', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ refresh_token: refreshToken })
                });
            } catch (e) {
            }
        }
        clearSession();
        showGuestUI();
    });

    // Сохраняем токены и пользователя в localStorage
    function saveSession(accessToken, refreshToken, user) {
        localStorage.setItem('access_token', accessToken);
        localStorage.setItem('refresh_token', refreshToken);
        localStorage.setItem('user', JSON.stringify(user));
    }

    // Очищаем сессию
    function clearSession() {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('user');
    }

    // Получаем сохранённого пользователя
    function getSavedUser() {
        const userStr = localStorage.getItem('user');
        return userStr ? JSON.parse(userStr) : null;
    }

    // Обновляем интерфейс: показываем блок пользователя, скрываем гостевые кнопки
    function showLoggedInUI(user) {
        if (guestButtons) guestButtons.style.display = 'none';
        if (userBlock) userBlock.style.display = 'block';
        if (userNameDisplay) userNameDisplay.textContent = user.username;
    }

    // Обновляем интерфейс для гостя
    function showGuestUI() {
        if (guestButtons) guestButtons.style.display = 'block';
        if (userBlock) userBlock.style.display = 'none';
        if (userNameDisplay) userNameDisplay.textContent = '';
    }

    // При загрузке страницы проверяем, есть ли сохранённая сессия
    function checkAuthOnLoad() {
        const user = getSavedUser();
        if (user) {
            showLoggedInUI(user);
        } else {
            showGuestUI();
        }
    }

    // Вызов при старте
    checkAuthOnLoad();
});