// courses.js

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
            <thead><tr><th>عنوان دوره</th><th>نام کوتاه</th><th>ترم</th><th>استاد</th><th>عملیات</th></tr></thead>
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
    document.querySelectorAll('.manage-course').forEach(btn => {
        btn.addEventListener('click', () => {
            const courseId = parseInt(btn.dataset.id);
            showCourseManage(courseId);
        });
    });
}

async function showCourseManage(courseId) {
    const listDiv = document.getElementById('courses-list');
    const editDiv = document.getElementById('course-edit-panel');
    if (!listDiv || !editDiv) return;
    listDiv.classList.add('hidden');
    editDiv.classList.remove('hidden');
    try {
        const response = await fetch(`/courses/${courseId}`);
        if (!response.ok) throw new Error('Failed to fetch course');
        const course = await response.json();
        populateCourseEditForm(course);
    } catch (err) {
        console.log(err);
        showToast('خطا در دریافت اطلاعات دوره', false);
        const backBtn = document.getElementById('back-to-courses-list');
        if (backBtn) backBtn.click();
    }
}

function populateCourseEditForm(course) {
    // Basic fields
    const courseIdField = document.getElementById('course-id');
    if (courseIdField) courseIdField.value = course.Id;
    const titleField = document.getElementById('course-title');
    if (titleField) titleField.value = course.Title || '';
    const shortNameField = document.getElementById('course-shortname');
    if (shortNameField) shortNameField.value = course.ShortName || '';
    const telegramField = document.getElementById('course-telegram');
    if (telegramField) telegramField.value = course.TelegramLink || '';
    const baleField = document.getElementById('course-bale');
    if (baleField) baleField.value = course.BaleLink || '';
    const queraField = document.getElementById('course-quera');
    if (queraField) queraField.value = course.QueraLink || '';
    const titleDisplay = document.getElementById('course-title-display');
    if (titleDisplay) titleDisplay.innerText = course.Title || 'بدون عنوان';

    // Semester and teacher dropdowns
    const semesterSelect = document.getElementById('course-semester');
    if (semesterSelect && course.Semester && course.Semester.Id) {
        semesterSelect.value = course.Semester.Id;
    }
    const teacherSelect = document.getElementById('course-teacher');
    if (teacherSelect && course.Teacher && course.Teacher.Id) {
        teacherSelect.value = course.Teacher.Id;
    }

    // Image preview with cache‑buster
    const previewImg = document.getElementById('course-image-preview');
    if (previewImg) {
        let imgUrl = (course.ImageUrl && course.ImageUrl !== '') ? course.ImageUrl : '/static/img/course-placeholder.jpg';
        // Add cache‑busting parameter to force fresh image from server
        const cacheBuster = Date.now();
        imgUrl = imgUrl + (imgUrl.includes('?') ? '&' : '?') + '_=' + cacheBuster;
        previewImg.src = imgUrl;
    }

    // Load other modules
    if (typeof loadCourseDescription === 'function') loadCourseDescription(course);
    if (typeof loadSlides === 'function') loadSlides(course.Id);
    if (typeof loadAssignments === 'function') loadAssignments(course.Id);
    if (typeof loadNotes === 'function') loadNotes(course.Id);
    if (typeof loadExams === 'function') loadExams(course.Id);
    if (typeof loadTAs === 'function') loadTAs(course.Id);
    if (course.Id && typeof loadBooks === 'function') {
        loadBooks(course.Id);
    } else if (typeof renderBooks === 'function') {
        renderBooks([]);
    }
}

async function loadSemesterOptions() {
    try {
        const response = await fetch('/semesters');
        const semesters = await response.json();
        const select = document.getElementById('course-semester');
        if (!select) return;
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
        if (!select) return;
        select.innerHTML = teachers.map(t => 
            `<option value="${t.Id}">${t.FirstName} ${t.LastName}</option>`
        ).join('');
    } catch (err) {
        console.error('Error loading teachers:', err);
    }
}

// ----- Delete Course logic -----
let pendingDeleteCourseIdForRemove = null;

function openDeleteCourseModal(courseId) {
    pendingDeleteCourseIdForRemove = courseId;
    const modal = document.getElementById('delete-course-modal');
    if (modal) modal.style.display = 'flex';
}

function closeDeleteCourseModal() {
    const modal = document.getElementById('delete-course-modal');
    if (modal) modal.style.display = 'none';
    pendingDeleteCourseIdForRemove = null;
}

async function confirmDeleteCourse() {
    if (!pendingDeleteCourseIdForRemove) return;
    try {
        const response = await fetch(`/courses/${pendingDeleteCourseIdForRemove}`, { method: 'DELETE' });
        if (!response.ok) throw new Error();
        showToast('دوره با موفقیت حذف شد', true);
        closeDeleteCourseModal();
        const backBtn = document.getElementById('back-to-courses-list');
        if (backBtn) backBtn.click();
    } catch (err) {
        showToast('خطا در حذف دوره', false);
        closeDeleteCourseModal();
    }
}

// ----- Add Course Modal -----
function openAddCourseModal() {
    const modal = document.getElementById('course-modal');
    if (!modal) return;
    const form = document.getElementById('course-modal-form');
    if (form) form.reset();
    const modalId = document.getElementById('course-modal-id');
    if (modalId) modalId.value = '0';
    const modalTitle = document.getElementById('course-modal-title');
    if (modalTitle) modalTitle.innerText = 'افزودن دوره جدید';
    loadCourseModalSemesterOptions();
    loadCourseModalTeacherOptions();
    modal.style.display = 'flex';
}

function closeCourseModal() {
    const modal = document.getElementById('course-modal');
    if (modal) modal.style.display = 'none';
}

async function loadCourseModalSemesterOptions() {
    try {
        const response = await fetch('/semesters');
        const semesters = await response.json();
        const select = document.getElementById('course-modal-semester');
        if (!select) return;
        select.innerHTML = semesters.map(s => 
            `<option value="${s.Id}">${s.Season === 'spring' ? 'بهار' : 'پاییز'} ${s.Year}</option>`
        ).join('');
    } catch (err) {
        console.error('Error loading semesters:', err);
    }
}

async function loadCourseModalTeacherOptions() {
    try {
        const response = await fetch('/teachers');
        const teachers = await response.json();
        const select = document.getElementById('course-modal-teacher');
        if (!select) return;
        select.innerHTML = teachers.map(t => 
            `<option value="${t.Id}">${t.FirstName} ${t.LastName}</option>`
        ).join('');
    } catch (err) {
        console.error('Error loading teachers:', err);
    }
}

// ----- Live image preview on file selection -----
function initImagePreview() {
    const fileInput = document.getElementById('course-image-file');
    const previewImg = document.getElementById('course-image-preview');
    if (!fileInput || !previewImg) return;
    if (fileInput.hasAttribute('data-preview-listener')) return;
    fileInput.setAttribute('data-preview-listener', 'true');
    fileInput.addEventListener('change', function(e) {
        const file = e.target.files[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = function(ev) {
                previewImg.src = ev.target.result;
            };
            reader.readAsDataURL(file);
        } else {
            // revert to stored image (with cache‑buster)
            const courseId = document.getElementById('course-id')?.value;
            if (courseId) {
                fetch(`/courses/${courseId}`)
                    .then(res => res.json())
                    .then(course => {
                        let imgUrl = (course.ImageUrl && course.ImageUrl !== '') ? course.ImageUrl : '/static/img/course-placeholder.jpg';
                        const cacheBuster = Date.now();
                        imgUrl = imgUrl + (imgUrl.includes('?') ? '&' : '?') + '_=' + cacheBuster;
                        previewImg.src = imgUrl;
                    })
                    .catch(() => {
                        previewImg.src = '/static/img/course-placeholder.jpg';
                    });
            } else {
                previewImg.src = '/static/img/course-placeholder.jpg';
            }
        }
    });
}

// ----- Initialization -----
function initCourses() {
    // Back button
    const backBtn = document.getElementById('back-to-courses-list');
    if (backBtn) {
        backBtn.addEventListener('click', () => {
            const listDiv = document.getElementById('courses-list');
            const editDiv = document.getElementById('course-edit-panel');
            if (listDiv) listDiv.classList.remove('hidden');
            if (editDiv) editDiv.classList.add('hidden');
            loadCourses();
        });
    }

    // Save basic info – do NOT reload course data (preserves blob preview)
    const saveBtn = document.getElementById('save-course-basic');
    if (saveBtn) {
        saveBtn.addEventListener('click', async () => {
            const courseId = document.getElementById('course-id')?.value;
            if (!courseId || courseId === '0') {
                showToast('شناسه دوره معتبر نیست', false);
                return;
            }
            const title = document.getElementById('course-title')?.value || '';
            const shortName = document.getElementById('course-shortname')?.value || '';
            const telegramLink = document.getElementById('course-telegram')?.value || '';
            const baleLink = document.getElementById('course-bale')?.value || '';
            const queraLink = document.getElementById('course-quera')?.value || '';
            const teacherId = parseInt(document.getElementById('course-teacher')?.value || '0');
            const semesterId = parseInt(document.getElementById('course-semester')?.value || '0');
            const imageFile = document.getElementById('course-image-file')?.files[0];

            if (!title || !shortName) {
                showToast('عنوان و نام کوتاه دوره الزامی است', false);
                return;
            }

            const formData = new FormData();
            formData.append('title', title);
            formData.append('shortName', shortName);
            if (telegramLink) formData.append('telegramLink', telegramLink);
            if (baleLink) formData.append('baleLink', baleLink);
            if (queraLink) formData.append('queraLink', queraLink);
            formData.append('teacherId', teacherId);
            formData.append('semesterId', semesterId);
            if (imageFile) formData.append('course_image', imageFile);

            const progressContainer = document.getElementById('course-edit-image-progress');
            const progressBar = document.getElementById('course-edit-image-progress-bar');

            const xhr = new XMLHttpRequest();
            xhr.open('PUT', `/courses/${courseId}/basic`, true);
            if (imageFile && progressContainer) {
                progressContainer.style.display = 'block';
                if (progressBar) progressBar.style.width = '0%';
                if (progressBar) progressBar.textContent = '0%';
                xhr.upload.addEventListener('progress', (ev) => {
                    if (ev.lengthComputable) {
                        const percent = (ev.loaded / ev.total) * 100;
                        if (progressBar) progressBar.style.width = percent + '%';
                        if (progressBar) progressBar.textContent = Math.round(percent) + '%';
                    }
                });
            }
            xhr.onload = () => {
                if (progressContainer) progressContainer.style.display = 'none';
                if (xhr.status === 200) {
                    showToast('اطلاعات پایه دوره ذخیره شد', true);
                    const titleDisplay = document.getElementById('course-title-display');
                    if (titleDisplay) titleDisplay.innerText = title;
                } else {
                    showToast('خطا در ذخیره اطلاعات پایه', false);
                }
            };
            xhr.onerror = () => {
                if (progressContainer) progressContainer.style.display = 'none';
                showToast('خطا در شبکه', false);
            };
            xhr.send(formData);
        });
    }

    // Delete course button (attach after edit panel appears)
    const attachDeleteButton = () => {
        const btn = document.getElementById('delete-course-btn');
        if (btn && !btn.hasAttribute('data-listener')) {
            btn.setAttribute('data-listener', 'true');
            btn.addEventListener('click', () => {
                const courseId = document.getElementById('course-id')?.value;
                if (courseId && courseId !== '0') {
                    openDeleteCourseModal(courseId);
                }
            });
        }
    };
    attachDeleteButton();

    // Delete modal events
    const confirmBtn = document.getElementById('confirm-delete-course-btn');
    if (confirmBtn) confirmBtn.addEventListener('click', confirmDeleteCourse);
    const cancelBtn = document.getElementById('cancel-delete-course-btn');
    if (cancelBtn) cancelBtn.addEventListener('click', closeDeleteCourseModal);
    const closeSpan = document.querySelector('.delete-course-close');
    if (closeSpan) closeSpan.addEventListener('click', closeDeleteCourseModal);
    window.addEventListener('click', (e) => {
        const modal = document.getElementById('delete-course-modal');
        if (e.target === modal) closeDeleteCourseModal();
    });

    // Add course modal events
    const addCourseBtn = document.getElementById('add-course-btn');
    if (addCourseBtn) addCourseBtn.addEventListener('click', openAddCourseModal);
    const courseModalClose = document.querySelector('.course-modal-close');
    if (courseModalClose) courseModalClose.addEventListener('click', closeCourseModal);
    const courseModalCancel = document.getElementById('course-modal-cancel');
    if (courseModalCancel) courseModalCancel.addEventListener('click', closeCourseModal);
    window.addEventListener('click', (e) => {
        const modal = document.getElementById('course-modal');
        if (e.target === modal) closeCourseModal();
    });

    // Course create form (add new course)
    const courseCreateForm = document.getElementById('course-modal-form');
    if (courseCreateForm && !courseCreateForm.hasAttribute('data-listener')) {
        courseCreateForm.setAttribute('data-listener', 'true');
        courseCreateForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const title = document.getElementById('course-modal-title-input')?.value || '';
            const shortName = document.getElementById('course-modal-shortname')?.value || '';
            const telegramLink = document.getElementById('course-modal-telegram')?.value || '';
            const baleLink = document.getElementById('course-modal-bale')?.value || '';
            const queraLink = document.getElementById('course-modal-quera')?.value || '';
            const teacherId = parseInt(document.getElementById('course-modal-teacher')?.value || '0');
            const semesterId = parseInt(document.getElementById('course-modal-semester')?.value || '0');
            const imageFile = document.getElementById('course-modal-image')?.files[0];

            if (!title || !shortName) {
                showToast('عنوان و نام کوتاه الزامی است', false);
                return;
            }

            const formData = new FormData();
            formData.append('title', title);
            formData.append('shortName', shortName);
            if (telegramLink) formData.append('telegramLink', telegramLink);
            if (baleLink) formData.append('baleLink', baleLink);
            if (queraLink) formData.append('queraLink', queraLink);
            formData.append('teacherId', teacherId);
            formData.append('semesterId', semesterId);
            if (imageFile) formData.append('course_image', imageFile);

            const progressContainer = document.getElementById('course-image-progress');
            const progressBar = document.getElementById('course-image-progress-bar');

            const xhr = new XMLHttpRequest();
            xhr.open('POST', '/courses', true);
            if (imageFile && progressContainer) {
                progressContainer.style.display = 'block';
                if (progressBar) progressBar.style.width = '0%';
                if (progressBar) progressBar.textContent = '0%';
                xhr.upload.addEventListener('progress', (ev) => {
                    if (ev.lengthComputable) {
                        const percent = (ev.loaded / ev.total) * 100;
                        if (progressBar) progressBar.style.width = percent + '%';
                        if (progressBar) progressBar.textContent = Math.round(percent) + '%';
                    }
                });
            }
            xhr.onload = () => {
                if (progressContainer) progressContainer.style.display = 'none';
                if (xhr.status === 201) {
                    showToast('دوره با موفقیت اضافه شد', true);
                    closeCourseModal();
                    loadCourses();
                } else {
                    showToast('خطا در ایجاد دوره', false);
                }
            };
            xhr.onerror = () => {
                if (progressContainer) progressContainer.style.display = 'none';
                showToast('خطا در شبکه', false);
            };
            xhr.send(formData);
        });
    }

    // Re-attach delete button when edit panel becomes visible
    const observer = new MutationObserver(() => {
        const editPanel = document.getElementById('course-edit-panel');
        if (editPanel && !editPanel.classList.contains('hidden')) {
            attachDeleteButton();
        }
    });
    observer.observe(document.body, { attributes: true, subtree: true, attributeFilter: ['class'] });

    // Initialize live image preview
    initImagePreview();
}