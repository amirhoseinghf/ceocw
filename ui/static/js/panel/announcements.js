let currentAnnouncementCourseId = 0;

function resetAnnouncementForm() {
    document.getElementById('announcement-id').value = '0';
    document.getElementById('announcement-title').value = '';
    document.getElementById('announcement-content').value = '';
}

async function loadAnnouncements(courseId) {
    currentAnnouncementCourseId = courseId;
    const container = document.getElementById('announcements-list');
    if (!container) return;
    container.innerHTML = '<div class="loading">در حال بارگذاری...</div>';
    try {
        const response = await fetch(`/courses/${courseId}/announcements`);
        if (!response.ok) throw new Error();
        const items = await response.json();
        renderAnnouncements(items || []);
    } catch (err) {
        container.innerHTML = '<div class="loading">خطا در بارگذاری اطلاعیه‌ها</div>';
    }
}

function renderAnnouncements(items) {
    const container = document.getElementById('announcements-list');
    if (!container) return;
    if (!items.length) {
        container.innerHTML = '<div class="loading">هنوز اطلاعیه‌ای ثبت نشده است.</div>';
        return;
    }
    container.innerHTML = `
        <table class="teachers-table">
            <thead><tr><th>عنوان</th><th>متن</th><th>عملیات</th></tr></thead>
            <tbody>
                ${items.map(item => `
                    <tr data-id="${item.Id}">
                        <td>${escapeHtml(item.Title || '')}</td>
                        <td>
                            <div>${escapeHtml(item.Content || '')}</div>
                            ${item.AuthorFirstName ? `<small class="muted-text">ثبت‌کننده: ${escapeHtml(item.AuthorFirstName)} ${escapeHtml(item.AuthorLastName || '')}</small>` : ''}
                        </td>
                        <td class="teacher-actions">
                            <button class="btn btn-edit edit-announcement" data-id="${item.Id}">ویرایش</button>
                            <button class="btn btn-delete delete-announcement" data-id="${item.Id}">حذف</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
    container.querySelectorAll('.edit-announcement').forEach(btn => {
        btn.addEventListener('click', () => {
            const row = btn.closest('tr');
            document.getElementById('announcement-id').value = btn.dataset.id;
            document.getElementById('announcement-title').value = row.children[0].textContent;
            document.getElementById('announcement-content').value = row.children[1].textContent;
        });
    });
    container.querySelectorAll('.delete-announcement').forEach(btn => {
        btn.addEventListener('click', async () => {
            if (!confirm('این اطلاعیه حذف شود؟')) return;
            const response = await fetch(`/announcements/${btn.dataset.id}`, { method: 'DELETE' });
            if (response.ok) {
                showToast('اطلاعیه حذف شد', true);
                resetAnnouncementForm();
                loadAnnouncements(currentAnnouncementCourseId);
            } else {
                showToast('خطا در حذف اطلاعیه', false);
            }
        });
    });
}

function initAnnouncements() {
    document.getElementById('clear-announcement')?.addEventListener('click', resetAnnouncementForm);
    document.getElementById('save-announcement')?.addEventListener('click', async () => {
        const courseId = Number(document.getElementById('course-id')?.value || currentAnnouncementCourseId || 0);
        const id = document.getElementById('announcement-id')?.value || '0';
        const title = document.getElementById('announcement-title')?.value.trim() || '';
        const content = document.getElementById('announcement-content')?.value.trim() || '';
        if (!courseId || !title || !content) {
            showToast('عنوان و متن اطلاعیه را وارد کنید', false);
            return;
        }
        const isEdit = id !== '0';
        const response = await fetch(isEdit ? `/announcements/${id}` : `/courses/${courseId}/announcements`, {
            method: isEdit ? 'PUT' : 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ title, content })
        });
        if (response.ok) {
            showToast(isEdit ? 'اطلاعیه ویرایش شد' : 'اطلاعیه ثبت شد', true);
            resetAnnouncementForm();
            loadAnnouncements(courseId);
        } else {
            showToast('خطا در ذخیره اطلاعیه', false);
        }
    });
}
