// semesters.js
let currentSemesterId = 0;

async function loadSemesters() {
    const container = document.getElementById('semesters-list');
    container.innerHTML = '<div class="loading">در حال بارگذاری...</div>';
    try {
        const response = await fetch('/semesters');
        if (!response.ok) throw new Error('Failed to fetch semesters');
        const semesters = await response.json();
        renderSemesters(semesters);
    } catch (err) {
        container.innerHTML = '<div class="loading">خطا در بارگذاری ترم‌ها</div>';
        console.error(err);
    }
}

function renderSemesters(semesters) {
    const container = document.getElementById('semesters-list');
    if (!semesters.length) {
        container.innerHTML = '<div class="loading">هیچ ترمی یافت نشد.</div>';
        return;
    }
    const html = `
        <table class="teachers-table">
            <thead><tr><th>ردیف</th><th>سال</th><th>فصل</th><th>عملیات</th></tr></thead>
            <tbody>
                ${semesters.map((sem, index) => `
                    <tr data-id="${sem.Id}">
                        <td>${index + 1}</td>
                        <td>${sem.Year}</td>
                        <td>${sem.Season === 'spring' ? 'بهار' : 'پاییز'}</td>
                        <td class="teacher-actions">
                            <button class="btn btn-edit edit-semester" data-id="${sem.Id}">✏️ ویرایش</button>
                            <button class="btn btn-delete delete-semester" data-id="${sem.Id}" data-name="${sem.Year} ${sem.Season === 'spring' ? 'بهار' : 'پاییز'}">🗑️ حذف</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
    container.innerHTML = html;
    document.querySelectorAll('.edit-semester').forEach(btn => {
        btn.addEventListener('click', () => openEditSemesterModal(parseInt(btn.dataset.id)));
    });
    document.querySelectorAll('.delete-semester').forEach(btn => {
        btn.addEventListener('click', () => {
            const id = parseInt(btn.dataset.id);
            const name = btn.dataset.name;
            showSemesterDeleteConfirm(id, name);
        });
    });
}

function openSemesterModal(isEdit = false, semesterData = null) {
    const modal = document.getElementById('semester-modal');
    const modalTitle = document.getElementById('semester-modal-title');
    if (isEdit && semesterData) {
        modalTitle.innerText = 'ویرایش ترم';
        document.getElementById('semester-id').value = semesterData.Id;
        document.getElementById('semester-year').value = semesterData.Year;
        document.getElementById('semester-season').value = semesterData.Season;
        currentSemesterId = semesterData.Id;
    } else {
        modalTitle.innerText = 'افزودن ترم جدید';
        document.getElementById('semester-form').reset();
        document.getElementById('semester-id').value = '0';
        document.getElementById('semester-year').value = getCurrentJalaliYear();
        document.getElementById('semester-season').value = 'spring';
        currentSemesterId = 0;
    }
    modal.style.display = 'flex';
}

function closeSemesterModal() {
    document.getElementById('semester-modal').style.display = 'none';
}

async function openEditSemesterModal(id) {
    try {
        const response = await fetch(`/semesters/${id}`);
        if (!response.ok) throw new Error();
        const semester = await response.json();
        openSemesterModal(true, semester);
    } catch (err) {
        showToast('خطا در دریافت اطلاعات ترم', false);
    }
}

// Delete logic for semesters
let pendingSemesterDeleteId = null;

function showSemesterDeleteConfirm(id, semesterName) {
    pendingSemesterDeleteId = id;
    const messageEl = document.getElementById('delete-semester-message');
    messageEl.innerHTML = `آیا از حذف ترم <strong>${escapeHtml(semesterName)}</strong> اطمینان دارید؟`;
    document.getElementById('delete-semester-modal').style.display = 'flex';
}

function closeSemesterDeleteModal() {
    document.getElementById('delete-semester-modal').style.display = 'none';
    pendingSemesterDeleteId = null;
}

async function confirmSemesterDelete() {
    if (!pendingSemesterDeleteId) return;
    try {
        const response = await fetch(`/semesters/${pendingSemesterDeleteId}`, { method: 'DELETE' });
        if (!response.ok) throw new Error('Delete failed');
        showToast('ترم با موفقیت حذف شد', true);
        closeSemesterDeleteModal();
        loadSemesters();
    } catch (err) {
        showToast('خطا در حذف ترم', false);
        closeSemesterDeleteModal();
    }
}

function initSemesters() {
    document.getElementById('semester-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const semesterData = {
            Year: parseInt(document.getElementById('semester-year').value),
            Season: document.getElementById('semester-season').value
        };
        const isEdit = currentSemesterId !== 0;
        const url = isEdit ? `/semesters/${currentSemesterId}` : '/semesters';
        const method = isEdit ? 'PUT' : 'POST';
        try {
            const response = await fetch(url, {
                method: method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(semesterData)
            });
            if (!response.ok) {
                const text = await response.text();
                throw new Error(text || 'Save failed');
            }
            showToast(isEdit ? 'ترم با موفقیت ویرایش شد' : 'ترم با موفقیت اضافه شد', true);
            closeSemesterModal();
            loadSemesters();
        } catch (err) {
            showToast(err.message.includes('سال') || err.message.includes('ترم') ? err.message : 'خطا در ذخیره اطلاعات ترم', false);
        }
    });
    document.getElementById('add-semester-btn').addEventListener('click', () => openSemesterModal(false));
    document.querySelector('.semester-close').addEventListener('click', closeSemesterModal);
    document.getElementById('semester-modal-cancel').addEventListener('click', closeSemesterModal);
    window.addEventListener('click', (e) => {
        if (e.target === document.getElementById('semester-modal')) closeSemesterModal();
    });
    document.getElementById('confirm-delete-semester-btn').addEventListener('click', confirmSemesterDelete);
    document.getElementById('cancel-delete-semester-btn').addEventListener('click', closeSemesterDeleteModal);
    document.querySelector('.delete-semester-close').addEventListener('click', closeSemesterDeleteModal);
    window.addEventListener('click', (e) => {
        if (e.target === document.getElementById('delete-semester-modal')) closeSemesterDeleteModal();
    });
}
