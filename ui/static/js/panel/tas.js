// tas.js

let currentTaId = 0;
let pendingDeleteTaId = null;
let pendingDeleteTACourseId = null;

// ------------------------------------------------------------
// Load and render course's attached TAs
// ------------------------------------------------------------
async function loadTAs(courseId) {
    const container = document.getElementById('tas-list');
    if (!container) return;
    container.innerHTML = '<div class="loading">در حال بارگذاری...</div>';
    try {
        const response = await fetch(`/courses/${courseId}/tas`);
        if (!response.ok) throw new Error();
        const tas = await response.json();
        renderTAs(tas);
    } catch (err) {
        console.error(err);
        container.innerHTML = '<div class="loading">خطا در بارگذاری TAها</div>';
    }
}

function renderTAs(tas) {
    const container = document.getElementById('tas-list');
    if (!container) return;
    if (!Array.isArray(tas)) {
        container.innerHTML = '<div class="loading">خطا در دریافت اطلاعات</div>';
        return;
    }
    if (tas.length === 0) {
        container.innerHTML = '<div class="loading">هیچ TAای برای این دوره ثبت نشده است.</div>';
        return;
    }
    const html = `
        <table class="teachers-table">
            <thead>
                <tr><th>تصویر</th><th>نام و نام خانوادگی</th><th>لینک‌ها</th><th>عملیات</th></tr>
            </thead>
            <tbody>
                ${tas.map(ta => `
                    <tr data-id="${ta.id}">
                        <td><img class="ta-avatar" src="${ta.imageUrl || '/static/img/teacher-placeholder.jpg'}" onerror="this.src='/static/img/teacher-placeholder.jpg'"></td>
                        <td>${escapeHtml(ta.firstName)} ${escapeHtml(ta.lastName)}</td>
                        <td>
                            <div class="ta-social-links">
                                ${ta.linkedin ? `<a href="${escapeHtml(ta.linkedin)}" target="_blank"><img src="/static/icons/small/linkedin.svg" alt="LinkedIn" class="social-icon"></a>` : ''}
                                ${ta.telegram ? `<a href="${escapeHtml(ta.telegram)}" target="_blank"><img src="/static/icons/small/telegram.svg" alt="Telegram" class="social-icon"></a>` : ''}
                                ${ta.instagram ? `<a href="${escapeHtml(ta.instagram)}" target="_blank"><img src="/static/icons/small/instagram.svg" alt="Instagram" class="social-icon"></a>` : ''}
                                ${ta.website ? `<a href="${escapeHtml(ta.website)}" target="_blank"><img src="/static/icons/small/website.svg" alt="Website" class="social-icon"></a>` : ''}
                                ${ta.github ? `<a href="${escapeHtml(ta.github)}" target="_blank"><img src="/static/icons/small/github.svg" alt="GitHub" class="social-icon"></a>` : ''}
                            </div>
                        </td>
                        <td class="teacher-actions">
                            <button class="btn btn-edit edit-ta" data-id="${ta.id}">✏️ ویرایش</button>
                            <button class="btn btn-delete delete-ta" data-id="${ta.id}">🗑️ حذف</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
    container.innerHTML = html;
    // Edit button
    document.querySelectorAll('.edit-ta').forEach(btn => {
        btn.addEventListener('click', () => openTAModal(parseInt(btn.dataset.id)));
    });
    // Delete button (opens modal)
    document.querySelectorAll('.delete-ta').forEach(btn => {
        btn.addEventListener('click', () => {
            const taId = parseInt(btn.dataset.id);
            openTADeleteModal(taId);
        });
    });
}

// ------------------------------------------------------------
// Delete modal (detach or permanently delete)
// ------------------------------------------------------------
function openTADeleteModal(taId) {
    pendingDeleteTaId = taId;
    pendingDeleteTACourseId = parseInt(document.getElementById('course-id').value);
    const taRow = document.querySelector(`.delete-ta[data-id="${taId}"]`).closest('tr');
    const taName = taRow.querySelector('td:nth-child(2)').innerText;
    document.getElementById('delete-ta-message').innerHTML = `آیا می‌خواهید TA <strong>${escapeHtml(taName)}</strong> را فقط از این دوره حذف کنید یا به طور کامل؟`;
    document.getElementById('delete-ta-modal').style.display = 'flex';
}

function closeTADeleteModal() {
    document.getElementById('delete-ta-modal').style.display = 'none';
    pendingDeleteTaId = null;
}

async function detachTA() {
    if (!pendingDeleteTaId || !pendingDeleteTACourseId) return;
    try {
        const response = await fetch(`/courses/${pendingDeleteTACourseId}/tas/${pendingDeleteTaId}`, { method: 'DELETE' });
        if (!response.ok) throw new Error();
        showToast('TA از این دوره حذف شد', true);
        closeTADeleteModal();
        loadTAs(pendingDeleteTACourseId);
    } catch (err) {
        showToast('خطا در حذف TA از دوره', false);
        closeTADeleteModal();
    }
}

async function deleteTAPermanently() {
    if (!pendingDeleteTaId) return;
    try {
        const response = await fetch(`/tas/${pendingDeleteTaId}`, { method: 'DELETE' });
        if (!response.ok) throw new Error();
        showToast('TA به طور کامل حذف شد', true);
        closeTADeleteModal();
        loadTAs(pendingDeleteTACourseId);
    } catch (err) {
        showToast('خطا در حذف کامل TA', false);
        closeTADeleteModal();
    }
}

// ------------------------------------------------------------
// Add/Edit TA modal
// ------------------------------------------------------------
async function openTAModal(taId = 0) {
    currentTaId = taId;
    const modal = document.getElementById('ta-modal');
    const form = document.getElementById('ta-form');
    form.reset();
    document.getElementById('ta-id').value = '0';
    document.getElementById('ta-image').value = '';
    if (taId) {
        try {
            const resp = await fetch(`/tas/${taId}`);
            const ta = await resp.json();
            document.getElementById('ta-id').value = ta.id;
            document.getElementById('ta-first-name').value = ta.firstName;
            document.getElementById('ta-last-name').value = ta.lastName;
            document.getElementById('ta-linkedin').value = ta.linkedin || '';
            document.getElementById('ta-telegram').value = ta.telegram || '';
            document.getElementById('ta-instagram').value = ta.instagram || '';
            document.getElementById('ta-website').value = ta.website || '';
            document.getElementById('ta-github').value = ta.github || '';
            document.getElementById('ta-modal-title').innerText = 'ویرایش TA';
        } catch (err) {
            showToast('خطا در دریافت اطلاعات TA', false);
            return;
        }
    } else {
        document.getElementById('ta-modal-title').innerText = 'افزودن TA جدید';
    }
    modal.style.display = 'flex';
}

function closeTAModal() {
    document.getElementById('ta-modal').style.display = 'none';
    currentTaId = 0;
}

// ------------------------------------------------------------
// Existing TA modal (attach existing TAs) – like books modal
// ------------------------------------------------------------
async function openExistingTAsModal() {
    const modal = document.getElementById('existing-tas-modal');
    modal.style.display = 'flex';
    await loadAllTAs();
}

async function loadAllTAs(searchTerm = '') {
    const container = document.getElementById('existing-tas-list');
    const courseId = document.getElementById('course-id').value;
    container.innerHTML = '<div class="loading">در حال بارگذاری...</div>';
    try {
        let url = '/tas';
        if (searchTerm) url += `?search=${encodeURIComponent(searchTerm)}`;
        const response = await fetch(url);
        if (!response.ok) throw new Error();
        const allTAs = await response.json();
        // Get currently attached TAs
        const attachedResponse = await fetch(`/courses/${courseId}/tas`);
        const attachedTAs = await attachedResponse.json();
        const attachedIds = new Set(attachedTAs.map(t => t.id));
        renderExistingTAs(allTAs, attachedIds);
    } catch (err) {
        console.error(err);
        container.innerHTML = '<div class="loading">خطا در بارگذاری TAها</div>';
    }
}

function renderExistingTAs(tas, attachedIds) {
    const container = document.getElementById('existing-tas-list');
    if (!tas.length) {
        container.innerHTML = '<div class="loading">هیچ TA ای یافت نشد.</div>';
        return;
    }
    const html = tas.map(ta => `
        <div class="ta-item" data-id="${ta.id}">
            <img src="${ta.imageUrl || '/static/img/teacher-placeholder.jpg'}" style="width:50px;height:50px;border-radius:50%;object-fit:cover;" onerror="this.src='/static/img/teacher-placeholder.jpg'">
            <div class="ta-name">${escapeHtml(ta.firstName)} ${escapeHtml(ta.lastName)}</div>
            ${attachedIds.has(ta.id) ? '<div class="ta-attached">✅ قبلاً اضافه شده</div>' : ''}
        </div>
    `).join('');
    container.innerHTML = html;
    // Attach click only for unattached TAs
    document.querySelectorAll('.ta-item').forEach(item => {
        const taId = parseInt(item.dataset.id);
        if (!attachedIds.has(taId)) {
            item.style.cursor = 'pointer';
            item.addEventListener('click', () => attachExistingTA(taId));
        } else {
            item.style.opacity = '0.6';
            item.style.cursor = 'default';
        }
    });
}

async function attachExistingTA(taId) {
    const courseId = document.getElementById('course-id').value;
    const response = await fetch(`/courses/${courseId}/tas/${taId}/attach`, { method: 'POST' });
    if (response.ok) {
        showToast('TA به دوره اضافه شد', true);
        closeExistingTAsModal();
        loadTAs(courseId);
    } else {
        showToast('خطا در افزودن TA', false);
    }
}

function closeExistingTAsModal() {
    document.getElementById('existing-tas-modal').style.display = 'none';
}

// ------------------------------------------------------------
// Initialisation
// ------------------------------------------------------------
function initTAs() {
    // Add new TA
    const addBtn = document.getElementById('add-ta-btn');
    if (addBtn) addBtn.addEventListener('click', () => openTAModal(0));

    // Add existing TA modal
    const addExistingBtn = document.getElementById('add-existing-ta-btn');
    if (addExistingBtn) {
        addExistingBtn.addEventListener('click', () => {
            const courseId = document.getElementById('course-id').value;
            if (courseId) openExistingTAsModal();
        });
    }

    // Close modals
    document.querySelector('.ta-close')?.addEventListener('click', closeTAModal);
    document.getElementById('ta-cancel')?.addEventListener('click', closeTAModal);
    window.addEventListener('click', (e) => {
        if (e.target === document.getElementById('ta-modal')) closeTAModal();
    });

    // Existing TA modal close events
    document.querySelector('.existing-tas-close')?.addEventListener('click', closeExistingTAsModal);
    document.getElementById('cancel-existing-tas')?.addEventListener('click', closeExistingTAsModal);
    window.addEventListener('click', (e) => {
        if (e.target === document.getElementById('existing-tas-modal')) closeExistingTAsModal();
    });

    // Search in existing TA modal
    const searchInput = document.getElementById('search-tas');
    if (searchInput) {
        let debounceTimer;
        searchInput.addEventListener('input', (e) => {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(() => {
                loadAllTAs(e.target.value);
            }, 300);
        });
    }

    // Delete modal buttons
    document.getElementById('detach-ta-btn')?.addEventListener('click', detachTA);
    document.getElementById('delete-permanently-ta-btn')?.addEventListener('click', deleteTAPermanently);
    document.getElementById('cancel-delete-ta-btn')?.addEventListener('click', closeTADeleteModal);
    document.querySelector('.delete-ta-close')?.addEventListener('click', closeTADeleteModal);
    window.addEventListener('click', (e) => {
        if (e.target === document.getElementById('delete-ta-modal')) closeTADeleteModal();
    });

    // Submit add/edit form
    const taForm = document.getElementById('ta-form');
    taForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const courseId = document.getElementById('course-id').value;
        const taId = document.getElementById('ta-id').value;
        const firstName = document.getElementById('ta-first-name').value;
        const lastName = document.getElementById('ta-last-name').value;
        const linkedin = document.getElementById('ta-linkedin').value;
        const telegram = document.getElementById('ta-telegram').value;
        const instagram = document.getElementById('ta-instagram').value;
        const website = document.getElementById('ta-website').value;
        const github = document.getElementById('ta-github').value;
        const imageFile = document.getElementById('ta-image').files[0];
 
        
        const formData = new FormData();
        formData.append('first_name', firstName);
        formData.append('last_name', lastName);
        if (linkedin) formData.append('linkedin', linkedin);
        if (telegram) formData.append('telegram', telegram);
        if (instagram) formData.append('instagram', instagram);
        if (website) formData.append('website', website);
        if (github) formData.append('github', github);
        if (imageFile) formData.append('ta_image', imageFile);

        let url, method;
        if (taId && taId !== '0') {
            url = `/tas/${taId}`;
            method = 'PUT';
        } else {
            url = `/courses/${courseId}/tas`;
            method = 'POST';
        }

        const progressContainer = document.getElementById('ta-progress-container');
        const progressBar = document.getElementById('ta-progress-bar');
        if (imageFile) {
            progressContainer.style.display = 'block';
            progressBar.style.width = '0%';
            progressBar.textContent = '0%';
        }

        const xhr = new XMLHttpRequest();
        xhr.open(method, url, true);
        if (imageFile) {
            xhr.upload.addEventListener('progress', (ev) => {
                if (ev.lengthComputable) {
                    const percent = (ev.loaded / ev.total) * 100;
                    progressBar.style.width = percent + '%';
                    progressBar.textContent = Math.round(percent) + '%';
                }
            });
        }
        xhr.onload = () => {
            if (progressContainer) progressContainer.style.display = 'none';
            if (xhr.status === 200 || xhr.status === 201) {
                showToast(taId ? 'TA ویرایش شد' : 'TA اضافه شد', true);
                closeTAModal();
                loadTAs(courseId);
            } else {
                showToast('خطا در ذخیره TA', false);
            }
        };
        xhr.onerror = () => {
            if (progressContainer) progressContainer.style.display = 'none';
            showToast('خطا در شبکه', false);
        };
        xhr.send(formData);
    });
}