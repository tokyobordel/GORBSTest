document.addEventListener('DOMContentLoaded', () => {
    const uploadBtn = document.getElementById('btnUpload');
    const uploadModal = document.getElementById('uploadModal');
    const uploadForm = document.getElementById('uploadForm');
    const fileInput = document.getElementById('fileInput');
    const fileList = document.getElementById('fileList');
    const uploadError = document.getElementById('uploadError');

    // Открытие
    uploadBtn.addEventListener('click', () => {
        window.openModal(uploadModal);
    });

    // Закрытие по крестику
    const closeBtn = uploadModal.querySelector('.close');
    if (closeBtn) {
        closeBtn.addEventListener('click', () => window.closeModal(uploadModal));
    }

    // Отображение выбранных файлов
    fileInput.addEventListener('change', () => {
        fileList.innerHTML = '';
        uploadError.textContent = '';

        const files = Array.from(fileInput.files);
        if (files.length > 3) {
            uploadError.textContent = 'Можно выбрать не более 3 изображений';
            fileInput.value = '';
            return;
        }

        const maxSize = 2 * 1024 * 1024;
        for (const file of files) {
            if (file.size > maxSize) {
                uploadError.textContent = `Файл "${file.name}" превышает 2 МБ`;
                fileInput.value = '';
                fileList.innerHTML = '';
                return;
            }
            if (!file.type.startsWith('image/')) {
                uploadError.textContent = `Файл "${file.name}" не является изображением`;
                fileInput.value = '';
                fileList.innerHTML = '';
                return;
            }
            const item = document.createElement('div');
            item.className = 'file-item';
            item.textContent = `${file.name} (${(file.size / 1024).toFixed(1)} КБ)`;
            fileList.appendChild(item);
        }
    });

    // Отправка
    uploadForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        uploadError.textContent = '';

        const files = fileInput.files;
        if (files.length === 0) {
            uploadError.textContent = 'Выберите хотя бы одно изображение';
            return;
        }

        const formData = new FormData();
        for (const file of files) {
            formData.append('images', file);
        }

        try {
            const response = await fetch('/upload', {
                method: 'POST',
                body: formData,
            });

            const data = await response.json();

            if (!response.ok || !data.success) {
                throw new Error(data.err_message || 'Ошибка загрузки');
            }

            window.closeModal(uploadModal);
            if (window.reloadFeed) window.reloadFeed();
        } catch (err) {
            uploadError.textContent = err.message;
        }
    });
});