// teachers.js
let currentEditId = 0;

async function loadTeachers() {
    const container = document.getElementById('teachers-list');
    container.innerHTML = '<div class="loading">در حال بارگذاری...</div>';
    try {
        const response = await fetch('/teachers');
        if (!response.ok) throw new Error('Failed to fetch');
        const teachers = await response.json();
        renderTeachers(teachers);
    } catch (err) {
        container.innerHTML = '<div class="loading">خطا در بارگذاری اساتید</div>';
        console.error(err);
    }
}

function renderTeachers(teachers) {
    const container = document.getElementById('teachers-list');
    if (!teachers.length) {
        container.innerHTML = '<div class="loading">هیچ استادی یافت نشد.</div>';
        return;
    }
    const html = `
        <table class="teachers-table">
            <thead><tr><th>تصویر</th><th>نام و نام خانوادگی</th><th>نام انگلیسی</th><th>لینک صفحه</th><th>عملیات</th></tr></thead>
            <tbody>
                ${teachers.map(teacher => `
                    <tr data-id="${teacher.Id}">
                        <td><img class="teacher-avatar" src="${teacher.ImageURL ? `${teacher.ImageURL}${teacher.ImageURL.includes('?') ? '&' : '?'}_=${Date.now()}` : '/static/img/teacher-placeholder.jpg'}" onerror="this.src='/static/img/teacher-placeholder.jpg'"></td>
                        <td>${escapeHtml(teacher.FirstName)} ${escapeHtml(teacher.LastName)}</td>
                        <td>${escapeHtml(teacher.FirstNameEnglish)} ${escapeHtml(teacher.LastNameEnglish)}</td>
                        <td><a href="${escapeHtml(teacher.PageURL)}" target="_blank">لینک صفحه</a></td>
                        <td class="teacher-actions">
                            <button class="btn btn-edit edit-teacher" data-id="${teacher.Id}">✏️ ویرایش</button>
                            <button class="btn btn-delete delete-teacher" data-id="${teacher.Id}" data-name="${escapeHtml(teacher.FirstName)} ${escapeHtml(teacher.LastName)}">🗑️ حذف</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
    container.innerHTML = html;
    document.querySelectorAll('.edit-teacher').forEach(btn => {
        btn.addEventListener('click', () => openEditModal(parseInt(btn.dataset.id)));
    });
    document.querySelectorAll('.delete-teacher').forEach(btn => {
        btn.addEventListener('click', () => {
            const id = parseInt(btn.dataset.id);
            const name = btn.dataset.name;
            showDeleteConfirm(id, name);
        });
    });
}

function openModal(isEdit = false, teacherData = null) {
    const modal = document.getElementById('teacher-modal');
    const modalTitle = document.getElementById('modal-title');
    const previewWrap = document.getElementById('teacher-image-preview-wrap');
    const previewImg = document.getElementById('teacher-image-preview');
    if (isEdit && teacherData) {
        modalTitle.innerText = 'ویرایش استاد';
        document.getElementById('teacher-id').value = teacherData.Id;
        document.getElementById('image-url').value = teacherData.ImageURL || '';
        document.getElementById('first-name').value = teacherData.FirstName || '';
        document.getElementById('last-name').value = teacherData.LastName || '';
        document.getElementById('first-name-en').value = teacherData.FirstNameEnglish || '';
        document.getElementById('last-name-en').value = teacherData.LastNameEnglish || '';
        document.getElementById('page-url').value = teacherData.PageURL || '';
        document.getElementById('teacher-image-file').value = '';
        if (teacherData.ImageURL) {
            previewImg.src = teacherData.ImageURL;
            previewWrap.style.display = 'block';
        } else {
            previewImg.src = '';
            previewWrap.style.display = 'none';
        }
        currentEditId = teacherData.Id;
    } else {
        modalTitle.innerText = 'افزودن استاد جدید';
        document.getElementById('teacher-form').reset();
        document.getElementById('teacher-id').value = '0';
        document.getElementById('image-url').value = '';
        previewImg.src = '';
        previewWrap.style.display = 'none';
        currentEditId = 0;
    }
    modal.style.display = 'flex';
}

function closeModal() {
    document.getElementById('teacher-modal').style.display = 'none';
}

async function openEditModal(id) {
    try {
        const response = await fetch(`/teachers/${id}`);
        if (!response.ok) throw new Error();
        const teacher = await response.json();
        openModal(true, teacher);
    } catch (err) {
        showToast('خطا در دریافت اطلاعات استاد', false);
    }
}

// Delete logic for teachers
let pendingDeleteId = null;

function showDeleteConfirm(id, teacherName) {
    pendingDeleteId = id;
    const messageEl = document.getElementById('delete-message');
    messageEl.innerHTML = `آیا از حذف استاد <strong>${escapeHtml(teacherName)}</strong> اطمینان دارید؟`;
    document.getElementById('delete-modal').style.display = 'flex';
}

function closeDeleteModal() {
    document.getElementById('delete-modal').style.display = 'none';
    pendingDeleteId = null;
}

async function confirmDelete() {
    if (!pendingDeleteId) return;
    try {
        const response = await fetch(`/teachers/${pendingDeleteId}`, { method: 'DELETE' });
        if (!response.ok) throw new Error('Delete failed');
        showToast('استاد با موفقیت حذف شد', true);
        closeDeleteModal();
        loadTeachers();
    } catch (err) {
        showToast('خطا در حذف استاد', false);
        closeDeleteModal();
    }
}

function initTeachers() {
    document.getElementById('teacher-image-file').addEventListener('change', function() {
        const file = this.files[0];
        const previewWrap = document.getElementById('teacher-image-preview-wrap');
        const previewImg = document.getElementById('teacher-image-preview');
        if (file) {
            const reader = new FileReader();
            reader.onload = function(e) {
                previewImg.src = e.target.result;
                previewWrap.style.display = 'block';
            };
            reader.readAsDataURL(file);
        }
    });

    document.getElementById('teacher-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const isEdit = currentEditId !== 0;
        const url = isEdit ? `/teachers/${currentEditId}` : '/teachers';
        const method = isEdit ? 'PUT' : 'POST';
        const formData = new FormData();
        formData.append('first_name', document.getElementById('first-name').value);
        formData.append('last_name', document.getElementById('last-name').value);
        formData.append('first_name_en', document.getElementById('first-name-en').value);
        formData.append('last_name_en', document.getElementById('last-name-en').value);
        formData.append('page_url', document.getElementById('page-url').value);
        formData.append('image_url', document.getElementById('image-url').value);
        const imageFile = document.getElementById('teacher-image-file').files[0];
        if (imageFile) formData.append('teacher_image', imageFile);
        try {
            const response = await fetch(url, { method: method, body: formData });
            if (!response.ok) throw new Error();
            showToast(isEdit ? 'استاد با موفقیت ویرایش شد' : 'استاد با موفقیت اضافه شد', true);
            closeModal();
            loadTeachers();
        } catch (err) {
            showToast('خطا در ذخیره اطلاعات', false);
        }
    });
    document.getElementById('add-teacher-btn').addEventListener('click', () => openModal(false));
    document.querySelector('.modal-close').addEventListener('click', closeModal);
    document.getElementById('modal-cancel').addEventListener('click', closeModal);
    window.addEventListener('click', (e) => { if (e.target === document.getElementById('teacher-modal')) closeModal(); });
    document.getElementById('confirm-delete-btn').addEventListener('click', confirmDelete);
    document.getElementById('cancel-delete-btn').addEventListener('click', closeDeleteModal);
    document.querySelector('.delete-close').addEventListener('click', closeDeleteModal);
    window.addEventListener('click', (e) => {
        if (e.target === document.getElementById('delete-modal')) closeDeleteModal();
    });
}
