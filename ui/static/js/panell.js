 // Tab switching
    document.addEventListener('DOMContentLoaded', () => {
        loadSemesterOptions();
        loadTeacherOptions();
        document.getElementById('add-semester-btn').addEventListener('click', () => openSemesterModal(false));
        const navItems = document.querySelectorAll('.nav-item');
        const tabs = {
            dashboard: document.getElementById('dashboard-content'),
            teachers: document.getElementById('teachers-content'),
            semesters: document.getElementById('semesters-content'),
            courses: document.getElementById('courses-content')
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
    if (tabId === 'courses') loadCourses();
}

        navItems.forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const tab = item.getAttribute('data-tab');
                if (tab) switchTab(tab);
            });
        });

    async function loadCourses() {
        const container = document.getElementById('courses-list');
        container.innerHTML = '<div class="loading">در حال بارگذاری...</div>';
        try {
            const response = await fetch('/courses');
            if (!response.ok) throw new Error('Failed to fetch courses');
            const courses = await response.json();
            renderCourses(courses);
        } catch (err) {
            container.innerHTML = '<div class="loading">خطا در بارگذاری دوره‌ها</div>';
            console.error(err);
        }
    }
    
    function renderCourses(courses) {
        const container = document.getElementById('courses-list');
        if (!courses.length) {
            container.innerHTML = '<div class="loading">هیچ دوره‌ای یافت نشد.</div>';
            return;
        }
        const html = `
            <table class="teachers-table">
                <thead>
                    <tr>
                        <th>عنوان دوره</th>
                        <th>نام کوتاه</th>
                        <th>ترم</th>
                        <th>استاد</th>
                        <th>عملیات</th>
                    </tr>
                </thead>
                <tbody>
                    ${courses.map(course => `
                        <tr data-id="${course.Id}">
                            <td>${escapeHtml(course.Title)}</td>
                            <td>${escapeHtml(course.ShortName)}</td>
                            <td>${escapeHtml(course.SemesterName)}</td>
                            <td>${escapeHtml(course.TeacherName)}</td>
                            <td class="teacher-actions">
                                <button class="btn btn-manage manage-course" data-id="${course.Id}">مدیریت</button>
                            </td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;
        container.innerHTML = html;
        // Attach manage events
        document.querySelectorAll('.manage-course').forEach(btn => {
            btn.addEventListener('click', () => {
                const courseId = parseInt(btn.dataset.id);
                showCourseManage(courseId); // we'll implement later
                showToast('مدیریت دوره در حال توسعه است', true); // placeholder
            });
        });
    }
    
    // Placeholder for now
    async function showCourseManage(courseId) {
        // Hide the courses list and show the edit panel
        const listDiv = document.getElementById('courses-list');
        const editDiv = document.getElementById('course-edit-panel');
        listDiv.classList.add('hidden');
        editDiv.classList.remove('hidden');
        
        // Load course data and populate the form
        try {
            const response = await fetch(`/courses/${courseId}`);
            if (!response.ok) throw new Error('Failed to fetch course');
            const course = await response.json();
            populateCourseEditForm(course);
        } catch (err) {
            showToast('خطا در دریافت اطلاعات دوره', false);
            // Go back to list on error
            document.getElementById('back-to-courses-list').click();
        }
    }

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

    function populateCourseEditForm(course) {
        document.getElementById('course-id').value = course.Id;
        document.getElementById('course-title').value = course.Title || '';
        document.getElementById('course-shortname').value = course.ShortName || '';
        document.getElementById('course-image').value = course.ImageUrl || '';
        document.getElementById('course-telegram').value = course.TelegramLink || '';
        document.getElementById('course-bale').value = course.BaleLink || '';
        document.getElementById('course-title-display').innerText = course.Title || 'بدون عنوان';
        
        // Set dropdown selections (assumes options are loaded)
        if (course.Semester && course.Semester.Id) {
            document.getElementById('course-semester').value = course.Semester.Id;
        }
        if (course.Teacher && course.Teacher.Id) {
            document.getElementById('course-teacher').value = course.Teacher.Id;
        }
    }

    async function loadSemesterOptions() {
        try {
            const response = await fetch('/semesters');
            const semesters = await response.json();
            const select = document.getElementById('course-semester');
            select.innerHTML = semesters.map(s => 
                `<option value="${s.Id}">${s.Season === 'spring' ? 'بهار' : 'پاییز'} ${s.Year}</option>`
            ).join('');
        } catch (err) {
            console.error('Error loading semesters:', err);
        }
    }

    async function loadTeacherOptions() {
        try {
            const response = await fetch('/teachers');
            const teachers = await response.json();
            const select = document.getElementById('course-teacher');
            select.innerHTML = teachers.map(t => 
                `<option value="${t.Id}">${t.FirstName} ${t.LastName}</option>`
            ).join('');
        } catch (err) {
            console.error('Error loading teachers:', err);
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
                            <button class="btn btn-delete delete-teacher" data-id="${teacher.Id}" data-name="${escapeHtml(teacher.FirstName)} ${escapeHtml(teacher.LastName)}">🗑️ حذف</button>
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

      document.querySelectorAll('.delete-teacher').forEach(btn => {
        btn.addEventListener('click', () => {
            const id = parseInt(btn.dataset.id);
            const name = btn.dataset.name;
            showDeleteConfirm(id, name);
        });
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
        semesterModal.style.display = 'none';

        const deleteModal = document.getElementById('delete-modal');
        deleteModal.style.display = 'none';
        
        const deleteSemesterModal = document.getElementById('delete-semester-modal');
        deleteSemesterModal.style.display = 'none';


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


        // Delete confirmation logic
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
        const response = await fetch(`/teachers/${pendingDeleteId}`, {
            method: 'DELETE'
        });
        if (!response.ok) throw new Error('Delete failed');
        showToast('استاد با موفقیت حذف شد', true);
        closeDeleteModal();
        loadTeachers(); // refresh list
    } catch (err) {
        showToast('خطا در حذف استاد', false);
        closeDeleteModal();
    }
}

// Bind modal events
document.getElementById('confirm-delete-btn').addEventListener('click', confirmDelete);
document.getElementById('cancel-delete-btn').addEventListener('click', closeDeleteModal);
document.querySelector('.delete-close').addEventListener('click', closeDeleteModal);
window.addEventListener('click', (e) => {
    const modal = document.getElementById('delete-modal');
    if (e.target === modal) closeDeleteModal();
});

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
        const response = await fetch(`/semesters/${pendingSemesterDeleteId}`, {
            method: 'DELETE'
        });
        if (!response.ok) throw new Error('Delete failed');
        showToast('ترم با موفقیت حذف شد', true);
        closeSemesterDeleteModal();
        loadSemesters(); // refresh list
    } catch (err) {
        showToast('خطا در حذف ترم', false);
        closeSemesterDeleteModal();
    }
}

// Bind semester delete modal events
document.getElementById('confirm-delete-semester-btn').addEventListener('click', confirmSemesterDelete);
document.getElementById('cancel-delete-semester-btn').addEventListener('click', closeSemesterDeleteModal);
document.querySelector('.delete-semester-close').addEventListener('click', closeSemesterDeleteModal);
window.addEventListener('click', (e) => {
    const modal = document.getElementById('delete-semester-modal');
    if (e.target === modal) closeSemesterDeleteModal();
});

document.getElementById('back-to-courses-list').addEventListener('click', () => {
    document.getElementById('courses-list').classList.remove('hidden');
    document.getElementById('course-edit-panel').classList.add('hidden');
    loadCourses(); // refresh the list in case of changes
});

document.getElementById('save-course-basic').addEventListener('click', async () => {
    const courseId = document.getElementById('course-id').value;
    if (!courseId || courseId === '0') {
        showToast('شناسه دوره معتبر نیست', false);
        return;
    }
    const basicData = {
        Title: document.getElementById('course-title').value,
        ShortName: document.getElementById('course-shortname').value,
        ImageUrl: document.getElementById('course-image').value,
        TelegramLink: document.getElementById('course-telegram').value,
        BaleLink: document.getElementById('course-bale').value,
        TeacherId: parseInt(document.getElementById('course-teacher').value),
        SemesterId: parseInt(document.getElementById('course-semester').value)
    };
    try {
        const response = await fetch(`/api/courses/${courseId}/basic`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(basicData)
        });
        if (!response.ok) throw new Error();
        showToast('اطلاعات پایه دوره ذخیره شد', true);
        // Update the display title
        document.getElementById('course-title-display').innerText = basicData.Title;
        // Optionally refresh the list later when going back
    } catch (err) {
        showToast('خطا در ذخیره اطلاعات', false);
    }
});

    });