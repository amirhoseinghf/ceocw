 // Tab switching
    document.addEventListener('DOMContentLoaded', () => {
        document.getElementById('add-semester-btn').addEventListener('click', () => openSemesterModal(false));
        const navItems = document.querySelectorAll('.nav-item');
        const tabs = {
            dashboard: document.getElementById('dashboard-content'),
            teachers: document.getElementById('teachers-content'),
            semesters: document.getElementById('semesters-content')
        };

        function switchTab(tabId) {
    // Only process tabs that actually exist in the DOM
    Object.keys(tabs).forEach(id => {
        const tabContent = tabs[id];
        const navItem = document.querySelector(`.nav-item[data-tab="${id}"]`);
        if (tabContent) tabContent.classList.remove('active');
        if (navItem) navItem.classList.remove('active');
    });
    
    const activeTab = tabs[tabId];
    const activeNav = document.querySelector(`.nav-item[data-tab="${tabId}"]`);
    if (activeTab) activeTab.classList.add('active');
    if (activeNav) activeNav.classList.add('active');

    if (tabId === 'teachers') loadTeachers();
    if (tabId === 'semesters') loadSemesters();
}

        navItems.forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const tab = item.getAttribute('data-tab');
                if (tab) switchTab(tab);
            });
        });

        // Load teachers when the tab becomes active
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


    function renderTeachers(teachers) {
    const container = document.getElementById('teachers-list');
    if (!teachers.length) {
        container.innerHTML = '<div class="loading">هیچ استادی یافت نشد.</div>';
        return;
    }
    const html = `
        <table class="teachers-table">
            <thead>
                <tr>
                    <th>تصویر</th>
                    <th>نام و نام خانوادگی</th>
                    <th>نام انگلیسی</th>
                    <th>لینک صفحه</th>
                    <th>عملیات</th>
                </tr>
            </thead>
            <tbody>
                ${teachers.map(teacher => `
                    <tr data-id="${teacher.Id}">
                        <td class="teacher-avatar-cell">
                            <img class="teacher-avatar" src="${teacher.ImageURL || 'https://via.placeholder.com/50'}" alt="avatar" onerror="this.src='https://via.placeholder.com/50'">
                        </td>
                        <td class="teacher-name">${escapeHtml(teacher.FirstName)} ${escapeHtml(teacher.LastName)}</td>
                        <td class="teacher-name-en">${escapeHtml(teacher.FirstNameEnglish)} ${escapeHtml(teacher.LastNameEnglish)}</td>
                        <td><a href="${escapeHtml(teacher.PageURL)}" class="teacher-page-link" target="_blank">لینک صفحه</a></td>
                        <td class="teacher-actions">
                            <button class="btn btn-edit edit-teacher" data-id="${teacher.Id}">✏️ ویرایش</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
    container.innerHTML = html;
    // Attach edit events
    document.querySelectorAll('.edit-teacher').forEach(btn => {
        btn.addEventListener('click', () => openEditModal(parseInt(btn.dataset.id)));
    });
}

    function renderSemesters(semesters) {
        const container = document.getElementById('semesters-list');
        if (!semesters.length) {
            container.innerHTML = '<div class="loading">هیچ ترمی یافت نشد.</div>';
            return;
        }
        const html = `
            <table class="teachers-table">
                <thead>
                    <tr>
                        <th>سال</th>
                        <th>فصل</th>
                        <th>عملیات</th>
                    </tr>
                </thead>
                <tbody>
                    ${semesters.map(sem => `
                        <tr data-id="${sem.Id}">
                            <td>${sem.Year}</td>
                            <td>${sem.Season === 'spring' ? 'بهار' : 'پاییز'}</td>
                            <td class="teacher-actions">
                                <button class="btn btn-edit edit-semester" data-id="${sem.Id}">✏️ ویرایش</button>
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
    }

        // Modal handling
        const modal = document.getElementById('teacher-modal');
        const modalTitle = document.getElementById('modal-title');
        const teacherForm = document.getElementById('teacher-form');
        const teacherIdInput = document.getElementById('teacher-id');
        const imageUrlInput = document.getElementById('image-url');
        const firstNameInput = document.getElementById('first-name');
        const lastNameInput = document.getElementById('last-name');
        const firstNameEnInput = document.getElementById('first-name-en');
        const lastNameEnInput = document.getElementById('last-name-en');
        const pageUrlInput = document.getElementById('page-url');
        let currentEditId = 0;
        modal.style.display = 'none';
        
        const semesterModal = document.getElementById('semester-modal');
        const semesterModalTitle = document.getElementById('semester-modal-title');
        const semesterForm = document.getElementById('semester-form');
        const semesterIdInput = document.getElementById('semester-id');
        const semesterYearInput = document.getElementById('semester-year');
        const semesterSeasonSelect = document.getElementById('semester-season');
        let currentSemesterId = 0;
        semesterModal.style.display = 'none'

        function openSemesterModal(isEdit = false, semesterData = null) {
    if (isEdit && semesterData) {
        semesterModalTitle.innerText = 'ویرایش ترم';
        semesterIdInput.value = semesterData.Id;
        semesterYearInput.value = semesterData.Year;
        semesterSeasonSelect.value = semesterData.Season;
        currentSemesterId = semesterData.Id;
    } else {
        semesterModalTitle.innerText = 'افزودن ترم جدید';
        semesterForm.reset();
        semesterIdInput.value = '0';
        semesterYearInput.value = new Date().getFullYear();
        semesterSeasonSelect.value = 'spring';
        currentSemesterId = 0;
    }
    semesterModal.style.display = 'flex';
}

function closeSemesterModal() {
    semesterModal.style.display = 'none';
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

semesterForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const semesterData = {
        Year: parseInt(semesterYearInput.value),
        Season: semesterSeasonSelect.value
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
        if (!response.ok) throw new Error();
        showToast(isEdit ? 'ترم با موفقیت ویرایش شد' : 'ترم با موفقیت اضافه شد', true);
        closeSemesterModal();
        loadSemesters(); // refresh list
    } catch (err) {
        showToast('خطا در ذخیره اطلاعات ترم', false);
    }
});

// Close modal events
document.querySelector('.semester-close').addEventListener('click', closeSemesterModal);
document.getElementById('semester-modal-cancel').addEventListener('click', closeSemesterModal);
window.addEventListener('click', (e) => { if (e.target === semesterModal) closeSemesterModal(); });
        
        
        function showToast(message, isSuccess) {
    var toast = document.getElementById('toast');
    toast.textContent = message;
    // Remove previous classes
    toast.classList.remove('success', 'error');
    // Add appropriate class
    toast.classList.add(isSuccess ? 'success' : 'error');
    toast.classList.add('show');
    setTimeout(function() {
        toast.classList.remove('show');
    }, 2800);
}

        function openModal(isEdit = false, teacherData = null) {
            if (isEdit && teacherData) {
                modalTitle.innerText = 'ویرایش استاد';
                teacherIdInput.value = teacherData.Id;
                imageUrlInput.value = teacherData.ImageURL || '';
                firstNameInput.value = teacherData.FirstName || '';
                lastNameInput.value = teacherData.LastName || '';
                firstNameEnInput.value = teacherData.FirstNameEnglish || '';
                lastNameEnInput.value = teacherData.LastNameEnglish || '';
                pageUrlInput.value = teacherData.PageURL || '';
                currentEditId = teacherData.Id;
            } else {
                modalTitle.innerText = 'افزودن استاد جدید';
                teacherForm.reset();
                teacherIdInput.value = '0';
                currentEditId = 0;
            }
            modal.style.display = 'flex';
        }

        function closeModal() {
            modal.style.display = 'none';
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

        function showToast(m,ok){var t=document.getElementById('toast');t.textContent=m;t.style.background=ok?'rgba(16,185,129,0.15)':'rgba(239,68,68,0.15)';t.style.borderColor=ok?'rgba(16,185,129,0.3)':'rgba(239,68,68,0.3)';t.style.color=ok?'#6ee7b7':'#fca5a5';t.classList.add('show');setTimeout(function(){t.classList.remove('show')},2800)}

        teacherForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const teacherData = {
                ImageURL: imageUrlInput.value,
                FirstName: firstNameInput.value,
                LastName: lastNameInput.value,
                FirstNameEnglish: firstNameEnInput.value,
                LastNameEnglish: lastNameEnInput.value,
                PageURL: pageUrlInput.value
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
                loadTeachers(); // refresh list
            } catch (err) {
                showToast('خطا در ذخیره اطلاعات', false);
            }
        });

        document.getElementById('add-teacher-btn').addEventListener('click', () => openModal(false));
        document.querySelector('.modal-close').addEventListener('click', closeModal);
        document.getElementById('modal-cancel').addEventListener('click', closeModal);
        window.addEventListener('click', (e) => { if (e.target === modal) closeModal(); });

        function escapeHtml(str) {
            if (!str) return '';
            return str.replace(/[&<>]/g, function(m) {
                if (m === '&') return '&amp;';
                if (m === '<') return '&lt;';
                if (m === '>') return '&gt;';
                return m;
            });
        }
    });