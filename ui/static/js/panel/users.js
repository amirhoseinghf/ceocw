// users.js

const ROLE_LABELS = {
    normal:   'کاربر عادی',
    ta:       'دستیار آموزشی',
    head_ta:  'سرپرست TA',
    admin:    'مدیر'
};
const ROLE_CLASSES = {
    normal:   'role-normal',
    ta:       'role-ta',
    head_ta:  'role-head-ta',
    admin:    'role-admin'
};

function roleLabel(role) {
    return ROLE_LABELS[role] || role;
}
function roleClass(role) {
    return ROLE_CLASSES[role] || 'role-normal';
}
function userInitials(firstName, lastName) {
    const f = (firstName || '').trim()[0] || '';
    const l = (lastName  || '').trim()[0] || '';
    return (f + l) || '?';
}

// ── Modal helpers ────────────────────────────────────────────────────────────
function openUserModal(isEdit, user) {
    const modal = document.getElementById('user-modal');
    const title = document.getElementById('user-modal-title');
    const pwdHint = document.getElementById('user-password-hint');
    const pwdInput = document.getElementById('user-password');

    if (isEdit && user) {
        title.textContent = 'ویرایش کاربر';
        document.getElementById('user-id').value = user.id;
        document.getElementById('user-first-name').value = user.firstName || '';
        document.getElementById('user-last-name').value  = user.lastName  || '';
        document.getElementById('user-email').value      = user.email     || '';
        document.getElementById('user-type').value       = user.userType  || 'normal';
        document.getElementById('user-is-active').checked = !!user.isActive;
        pwdInput.value    = '';
        pwdInput.required = false;
        pwdHint.style.display = '';
    } else {
        title.textContent = 'افزودن کاربر جدید';
        document.getElementById('user-form').reset();
        document.getElementById('user-id').value = '0';
        document.getElementById('user-is-active').checked = true;
        pwdInput.required = true;
        pwdHint.style.display = 'none';
    }
    modal.style.display = 'flex';
}

function closeUserModal() {
    document.getElementById('user-modal').style.display = 'none';
}

// ── Load & render ────────────────────────────────────────────────────────────
async function loadUsers() {
    if (!isAdminUser()) return;
    const container = document.getElementById('users-list');
    if (!container) return;
    container.innerHTML = '<div class="loading">در حال بارگذاری...</div>';
    try {
        const res = await fetch('/users');
        if (!res.ok) throw new Error();
        const users = await res.json();
        renderUsers(users);
    } catch {
        container.innerHTML = '<div class="loading">خطا در بارگذاری کاربران</div>';
    }
}

function renderUsers(users) {
    const container = document.getElementById('users-list');
    if (!container) return;
    window._panelUsersById = new Map((users || []).map(u => [Number(u.id), u]));

    if (!users || !users.length) {
        container.innerHTML = '<div class="loading">هیچ کاربری ثبت نشده است.</div>';
        return;
    }

    const rows = users.map(u => `
        <tr data-id="${u.id}">
            <td>
                <div class="user-avatar-cell">
                    <div class="user-avatar-initials">${escapeHtml(userInitials(u.firstName, u.lastName))}</div>
                    <div class="user-name-block">
                        <span class="user-fullname">${escapeHtml(u.firstName || '')} ${escapeHtml(u.lastName || '')}</span>
                        <span class="user-email" dir="ltr">${escapeHtml(u.email || '')}</span>
                    </div>
                </div>
            </td>
            <td><span class="role-badge ${roleClass(u.userType)}">${roleLabel(u.userType)}</span></td>
            <td><span class="status-badge ${u.isActive ? 'status-active' : 'status-inactive'}">${u.isActive ? 'فعال' : 'غیرفعال'}</span></td>
            <td class="teacher-actions">
                <button class="btn btn-edit edit-user-btn" data-id="${u.id}">✏️ ویرایش</button>
                <button class="btn ${u.isActive ? 'btn-delete' : 'btn-success'} toggle-user-btn"
                        data-id="${u.id}" data-active="${u.isActive ? '1' : '0'}">
                    ${u.isActive ? '🚫 غیرفعال' : '✅ فعال‌سازی'}
                </button>
            </td>
        </tr>
    `).join('');

    container.innerHTML = `
        <div class="edit-section" style="padding:0; overflow:hidden;">
            <table class="teachers-table">
                <thead>
                    <tr>
                        <th>کاربر</th>
                        <th>نقش</th>
                        <th>وضعیت</th>
                        <th>عملیات</th>
                    </tr>
                </thead>
                <tbody>${rows}</tbody>
            </table>
        </div>
    `;

    container.querySelectorAll('.edit-user-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const user = window._panelUsersById.get(Number(btn.dataset.id));
            openUserModal(true, user);
        });
    });
    container.querySelectorAll('.toggle-user-btn').forEach(btn => {
        btn.addEventListener('click', () => toggleUserActive(Number(btn.dataset.id), btn.dataset.active === '1'));
    });
}

// ── Toggle active / inactive ─────────────────────────────────────────────────
async function toggleUserActive(id, currentlyActive) {
    const action = currentlyActive ? 'غیرفعال کردن' : 'فعال‌سازی';
    if (!confirm(`آیا از ${action} این کاربر اطمینان دارید؟`)) return;
    try {
        const res = await fetch(`/users/${id}`, { method: 'DELETE' });
        if (!res.ok) throw new Error();
        showToast(currentlyActive ? 'کاربر غیرفعال شد' : 'کاربر فعال شد', true);
        loadUsers();
    } catch {
        showToast('خطا در تغییر وضعیت کاربر', false);
    }
}

// ── Init ─────────────────────────────────────────────────────────────────────
function initUsers() {
    if (!isAdminUser()) return;
    const form = document.getElementById('user-form');
    if (!form) return;

    // "Add user" button
    document.getElementById('add-user-btn')?.addEventListener('click', () => openUserModal(false));

    // Modal close
    document.querySelector('.user-modal-close')?.addEventListener('click', closeUserModal);
    document.getElementById('user-modal-cancel')?.addEventListener('click', closeUserModal);
    window.addEventListener('click', e => {
        if (e.target === document.getElementById('user-modal')) closeUserModal();
    });

    // Form submit
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const id = Number(document.getElementById('user-id').value || '0');
        const payload = {
            firstName: document.getElementById('user-first-name').value.trim(),
            lastName:  document.getElementById('user-last-name').value.trim(),
            email:     document.getElementById('user-email').value.trim(),
            password:  document.getElementById('user-password').value,
            userType:  document.getElementById('user-type').value,
            isActive:  document.getElementById('user-is-active').checked
        };
        if (!payload.firstName || !payload.lastName || !payload.email) {
            showToast('نام، نام خانوادگی و ایمیل الزامی است', false);
            return;
        }
        if (!id && !payload.password) {
            showToast('رمز عبور برای کاربر جدید الزامی است', false);
            return;
        }
        try {
            const res = await fetch(id ? `/users/${id}` : '/users', {
                method: id ? 'PUT' : 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            if (!res.ok) {
                const text = await res.text();
                throw new Error(text || 'خطا در ذخیره');
            }
            showToast(id ? 'کاربر ویرایش شد' : 'کاربر اضافه شد', true);
            closeUserModal();
            loadUsers();
        } catch (err) {
            showToast(err.message || 'خطا در ذخیره کاربر', false);
        }
    });
}
