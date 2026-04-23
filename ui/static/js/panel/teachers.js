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
                        <td><img class="teacher-avatar" src="${teacher.ImageURL || 'https://via.placeholder.com/50'}" onerror="this.src='https://via.placeholder.com/50'"></td>
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
    if (isEdit && teacherData) {
        modalTitle.innerText = 'ویرایش استاد';
        document.getElementById('teacher-id').value = teacherData.Id;
        document.getElementById('image-url').value = teacherData.ImageURL || '';
        document.getElementById('first-name').value = teacherData.FirstName || '';
        document.getElementById('last-name').value = teacherData.LastName || '';
        document.getElementById('first-name-en').value = teacherData.FirstNameEnglish || '';
        document.getElementById('last-name-en').value = teacherData.LastNameEnglish || '';
        document.getElementById('page-url').value = teacherData.PageURL || '';
        currentEditId = teacherData.Id;
    } else {
        modalTitle.innerText = 'افزودن استاد جدید';
        document.getElementById('teacher-form').reset();
        document.getElementById('teacher-id').value = '0';
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
    document.getElementById('teacher-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const teacherData = {
            ImageURL: document.getElementById('image-url').value,
            FirstName: document.getElementById('first-name').value,
            LastName: document.getElementById('last-name').value,
            FirstNameEnglish: document.getElementById('first-name-en').value,
            LastNameEnglish: document.getElementById('last-name-en').value,
            PageURL: document.getElementById('page-url').value
        };
        const isEdit = currentEditId !== 0;
        const url = isEdit ? `/teachers/${currentEditId}` : '/teachers';
        const method = isEdit ? 'PUT' : 'POST';
        try {
            const response = await fetch(url, {
                method: method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(teacherData)
            });
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