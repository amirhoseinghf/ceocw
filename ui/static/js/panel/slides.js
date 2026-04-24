// slides.js
async function loadSlides(courseId) {
    const container = document.getElementById('slides-list');
    if (!container) return;
    container.innerHTML = '<div class="loading">در حال بارگذاری...</div>';
    try {
        const response = await fetch(`/courses/${courseId}/slides`);
        if (!response.ok) throw new Error();
        const slides = await response.json();
        renderSlides(slides);
    } catch (err) {
        console.error(err);
        container.innerHTML = '<div class="loading">خطا در بارگذاری اسلایدها</div>';
    }
}

function renderSlides(slides) {
    const container = document.getElementById('slides-list');
    if (!container) return;
    if (!Array.isArray(slides)) {
        container.innerHTML = '<div class="loading">خطا در دریافت اطلاعات</div>';
        return;
    }
    if (slides.length === 0) {
        container.innerHTML = '<div class="loading">هیچ اسلایدی ثبت نشده است.</div>';
        return;
    }
    const html = `
        <table class="teachers-table slides-table">
            <thead>
                <tr><th>عنوان</th><th>فایل</th><th>عملیات</th></tr>
            </thead>
            <tbody>
                ${slides.map(slide => `
                    <tr data-id="${slide.Id}">
                        <td class="slide-title">${escapeHtml(slide.Title)}</td>
                        <td><a href="${escapeHtml(slide.FileName)}" target="_blank">دانلود</a></td>
                        <td class="slide-actions">
                            <button class="btn btn-delete delete-slide" data-id="${slide.Id}">🗑️ حذف</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
    container.innerHTML = html;
    document.querySelectorAll('.delete-slide').forEach(btn => {
        btn.addEventListener('click', async () => {
            if (confirm('آیا از حذف این اسلاید اطمینان دارید؟')) {
                await deleteSlide(parseInt(btn.dataset.id));
                const courseId = document.getElementById('course-id').value;
                if (courseId) loadSlides(courseId);
            }
        });
    });
}

async function deleteSlide(slideId) {
    const response = await fetch(`/slides/${slideId}`, { method: 'DELETE' });
    if (response.ok) {
        showToast('اسلاید حذف شد', true);
        return true;
    } else {
        showToast('خطا در حذف اسلاید', false);
        return false;
    }
}

function openSlideModal() {
    const modal = document.getElementById('slide-modal');
    document.getElementById('slide-form').reset();
    document.getElementById('slide-id').value = '0';
    document.getElementById('slide-file').value = '';
    document.getElementById('slide-modal-title').innerText = 'افزودن اسلاید جدید';
    modal.style.display = 'flex';
}

function closeSlideModal() {
    document.getElementById('slide-modal').style.display = 'none';
}

function initSlides() {
    const addBtn = document.getElementById('add-slide-btn');
    if (addBtn) addBtn.addEventListener('click', openSlideModal);
    document.querySelector('.slide-close')?.addEventListener('click', closeSlideModal);
    document.getElementById('slide-cancel')?.addEventListener('click', closeSlideModal);
    window.addEventListener('click', (e) => {
        if (e.target === document.getElementById('slide-modal')) closeSlideModal();
    });

    const slideForm = document.getElementById('slide-form');
    slideForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const courseId = document.getElementById('course-id').value;
        const title = document.getElementById('slide-title').value;
        const fileInput = document.getElementById('slide-file');
        const file = fileInput.files[0];
        if (!title || !file) {
            showToast('عنوان و فایل اسلاید الزامی است', false);
            return;
        }
        const formData = new FormData();
        formData.append('title', title);
        formData.append('slide_file', file);

        // Progress bar
        const progressContainer = document.getElementById('slide-progress-container');
        const progressBar = document.getElementById('slide-progress-bar');
        progressContainer.style.display = 'block';
        progressBar.style.width = '0%';
        progressBar.textContent = '0%';

        const xhr = new XMLHttpRequest();
        xhr.open('POST', `/courses/${courseId}/slides`, true);
        xhr.upload.addEventListener('progress', (ev) => {
            if (ev.lengthComputable) {
                const percent = (ev.loaded / ev.total) * 100;
                progressBar.style.width = percent + '%';
                progressBar.textContent = Math.round(percent) + '%';
            }
        });
        xhr.onload = () => {
            progressContainer.style.display = 'none';
            if (xhr.status === 201) {
                showToast('اسلاید اضافه شد', true);
                closeSlideModal();
                loadSlides(courseId);
            } else {
                showToast('خطا در ذخیره اسلاید', false);
            }
        };
        xhr.onerror = () => {
            progressContainer.style.display = 'none';
            showToast('خطا در شبکه', false);
        };
        xhr.send(formData);
    });
}