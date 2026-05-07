async function loadCourseUsers(courseId) {
    if (!isAdminUser()) return;
    const container = document.getElementById('course-users-list');
    if (!container) return;
    container.innerHTML = '<div class="loading">در حال بارگذاری...</div>';
    try {
        const [assignedResponse, usersResponse] = await Promise.all([
            fetch(`/courses/${courseId}/users`),
            fetch('/users/assignable')
        ]);
        if (!assignedResponse.ok || !usersResponse.ok) throw new Error();
        const assigned = await assignedResponse.json();
        const users = await usersResponse.json();
        renderCourseUsers(assigned);
        fillAssignableUsers(users, assigned);
    } catch (err) {
        container.innerHTML = '<div class="loading">خطا در بارگذاری کاربران دوره</div>';
    }
}

function fillAssignableUsers(users, assigned) {
    const select = document.getElementById('course-user-select');
    if (!select) return;
    const assignedIds = new Set((assigned || []).map(user => Number(user.userId || user.UserID)));
    const available = (users || []).filter(user => !assignedIds.has(Number(user.id)));
    if (!available.length) {
        select.innerHTML = '<option value="">کاربری برای تخصیص وجود ندارد</option>';
        return;
    }
    select.innerHTML = available.map(user => `
        <option value="${user.id}" data-role="${user.userType}">${escapeHtml(user.firstName)} ${escapeHtml(user.lastName)} - ${roleLabel(user.userType)}</option>
    `).join('');
    syncCourseUserRole();
}

function renderCourseUsers(users) {
    const container = document.getElementById('course-users-list');
    if (!container) return;
    if (!users.length) {
        container.innerHTML = '<div class="loading">هنوز کاربری به این دوره تخصیص داده نشده است.</div>';
        return;
    }
    container.innerHTML = `
        <table class="teachers-table">
            <thead>
                <tr><th>نام</th><th>ایمیل</th><th>نقش در دوره</th><th>عملیات</th></tr>
            </thead>
            <tbody>
                ${users.map(user => `
                    <tr>
                        <td>${escapeHtml(user.firstName || user.FirstName)} ${escapeHtml(user.lastName || user.LastName)}</td>
                        <td dir="ltr">${escapeHtml(user.email || user.Email)}</td>
                        <td>${roleLabel(user.role || user.Role)}</td>
                        <td class="teacher-actions">
                            <button class="btn btn-delete remove-course-user" data-user-id="${user.userId || user.UserID}">حذف از دوره</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
    document.querySelectorAll('.remove-course-user').forEach(btn => {
        btn.addEventListener('click', () => removeCourseUser(Number(btn.dataset.userId)));
    });
}

async function assignCourseUser() {
    const courseId = Number(document.getElementById('course-id')?.value || '0');
    const userId = Number(document.getElementById('course-user-select')?.value || '0');
    const roleSelect = document.getElementById('course-user-role');
    const selectedRole = document.getElementById('course-user-select')?.selectedOptions?.[0]?.dataset?.role;
    const role = selectedRole || roleSelect?.value || 'ta';
    if (!courseId || !userId) {
        showToast('کاربر و دوره را انتخاب کنید', false);
        return;
    }
    try {
        const response = await fetch(`/courses/${courseId}/users`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ userId: userId, role: role })
        });
        if (!response.ok) throw new Error(await response.text());
        showToast('کاربر به دوره اضافه شد', true);
        loadCourseUsers(courseId);
    } catch (err) {
        showToast(err.message || 'خطا در تخصیص کاربر', false);
    }
}

async function removeCourseUser(userId) {
    const courseId = Number(document.getElementById('course-id')?.value || '0');
    if (!courseId || !userId) return;
    try {
        const response = await fetch(`/courses/${courseId}/users/${userId}`, { method: 'DELETE' });
        if (!response.ok) throw new Error();
        showToast('کاربر از دوره حذف شد', true);
        loadCourseUsers(courseId);
    } catch (err) {
        showToast('خطا در حذف کاربر از دوره', false);
    }
}

function initCourseUsers() {
    if (!isAdminUser()) return;
    document.getElementById('assign-course-user-btn')?.addEventListener('click', assignCourseUser);
    document.getElementById('course-user-select')?.addEventListener('change', syncCourseUserRole);
}

function syncCourseUserRole() {
    const selectedRole = document.getElementById('course-user-select')?.selectedOptions?.[0]?.dataset?.role;
    const roleSelect = document.getElementById('course-user-role');
    if (selectedRole && roleSelect) roleSelect.value = selectedRole;
}
