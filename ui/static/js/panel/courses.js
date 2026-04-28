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
    listDiv.classList.add('hidden');
    editDiv.classList.remove('hidden');
    try {
        const response = await fetch(`/courses/${courseId}`);
        if (!response.ok) throw new Error('Failed to fetch course');
        const course = await response.json();
        populateCourseEditForm(course);
    } catch (err) {
        showToast('خطا در دریافت اطلاعات دوره', false);
        document.getElementById('back-to-courses-list').click();
    }
}

function populateCourseEditForm(course) {
    document.getElementById('course-id').value = course.Id;
    document.getElementById('course-title').value = course.Title || '';
    document.getElementById('course-shortname').value = course.ShortName || '';
    document.getElementById('current-course-image').innerText = course.ImageUrl || 'بدون تصویر';
    document.getElementById('course-telegram').value = course.TelegramLink || '';
    document.getElementById('course-bale').value = course.BaleLink || '';
    document.getElementById('course-title-display').innerText = course.Title || 'بدون عنوان';
    document.getElementById('course-quera').value = course.QueraLink || '';

    if (course.Semester && course.Semester.Id) {
        document.getElementById('course-semester').value = course.Semester.Id;
    }
    if (course.Teacher && course.Teacher.Id) {
        document.getElementById('course-teacher').value = course.Teacher.Id;
    }

    loadCourseDescription(course);
    loadSlides(course.Id);
    loadAssignments(course.Id);
    loadNotes(course.Id);
    loadExams(course.Id);
    loadTAs(course.Id);
    
    if (course.Id) {
        loadBooks(course.Id)
    } else {
        renderBooks([]);
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
        // Go back to list and refresh
        document.getElementById('back-to-courses-list').click();
    } catch (err) {
        showToast('خطا در حذف دوره', false);
        closeDeleteCourseModal();
    }
}

// ----- Add Course Modal -----
function openAddCourseModal() {
    const modal = document.getElementById('course-modal');
    const form = document.getElementById('course-modal-form');
    form.reset();
    document.getElementById('course-modal-id').value = '0';
    document.getElementById('course-modal-title').innerText = 'افزودن دوره جدید';
    loadCourseModalSemesterOptions();
    loadCourseModalTeacherOptions();
    modal.style.display = 'flex';
}

function closeCourseModal() {
    document.getElementById('course-modal').style.display = 'none';
}

async function loadCourseModalSemesterOptions() {
    try {
        const response = await fetch('/semesters');
        const semesters = await response.json();
        const select = document.getElementById('course-modal-semester');
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
        select.innerHTML = teachers.map(t => 
            `<option value="${t.Id}">${t.FirstName} ${t.LastName}</option>`
        ).join('');
    } catch (err) {
        console.error('Error loading teachers:', err);
    }
}

// ----- Initialization -----
function initCourses() {
    // Back button
    const backBtn = document.getElementById('back-to-courses-list');
    if (backBtn) {
        backBtn.addEventListener('click', () => {
            document.getElementById('courses-list').classList.remove('hidden');
            document.getElementById('course-edit-panel').classList.add('hidden');
            loadCourses();
        });
    }

    // Save basic info (already implemented)
        document.getElementById('save-course-basic').addEventListener('click', async () => {
        const courseId = document.getElementById('course-id').value;
        if (!courseId || courseId === '0') {
            showToast('شناسه دوره معتبر نیست', false);
            return;
        }

        const title = document.getElementById('course-title').value;
        const shortName = document.getElementById('course-shortname').value;
        const telegramLink = document.getElementById('course-telegram').value;
        const baleLink = document.getElementById('course-bale').value;
        const queraLink = document.getElementById('course-quera').value;
        const teacherId = parseInt(document.getElementById('course-teacher').value);
        const semesterId = parseInt(document.getElementById('course-semester').value);
        const imageFile = document.getElementById('course-image-file').files[0];

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
        if (imageFile) {
            progressContainer.style.display = 'block';
            progressBar.style.width = '0%';
            progressBar.textContent = '0%';
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
            if (xhr.status === 200) {
                showToast('اطلاعات پایه دوره ذخیره شد', true);
                document.getElementById('course-title-display').innerText = title;
                // Optionally update the displayed current image text
                if (imageFile) {
                    // After successful upload, we could fetch the new image URL, but we'll just show a placeholder.
                    document.getElementById('current-course-image').innerText = 'تصویر جدید آپلود شد';
                }
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

    // Delete course button (inside edit panel, may not exist initially – use event delegation or attach later)
    // We'll attach when the edit panel becomes visible? Instead, we can attach listener once and check existence each time.
    // Better: use event delegation or attach inside showCourseManage. But easiest: attach globally but check if element exists.
    const deleteBtn = document.getElementById('delete-course-btn');
    if (deleteBtn) {
        deleteBtn.addEventListener('click', () => {
            const courseId = document.getElementById('course-id').value;
            if (courseId && courseId !== '0') {
                openDeleteCourseModal(courseId);
            }
        });
    } else {
        // If button not present yet (e.g., DOM loaded but edit panel hidden), we can attach later when edit panel is shown.
        // We'll use MutationObserver or simply re-attach inside showCourseManage. Simpler: attach in showCourseManage after panel appears.
        // For robustness, we'll also check inside showCourseManage after populating.
    }

    // Delete modal confirm/cancel events
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

    // Course create form
    const courseCreateForm = document.getElementById('course-modal-form');
    if (courseCreateForm && !courseCreateForm.hasAttribute('data-listener')) {
        courseCreateForm.setAttribute('data-listener', 'true');
        courseCreateForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const title = document.getElementById('course-modal-title-input').value;
            const shortName = document.getElementById('course-modal-shortname').value;
            const telegramLink = document.getElementById('course-modal-telegram').value;
            const baleLink = document.getElementById('course-modal-bale').value;
            const queraLink = document.getElementById('course-modal-quera').value;
            const teacherId = parseInt(document.getElementById('course-modal-teacher').value);
            const semesterId = parseInt(document.getElementById('course-modal-semester').value);
            const imageFile = document.getElementById('course-modal-image').files[0];
        
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
            if (imageFile) {
                progressContainer.style.display = 'block';
                progressBar.style.width = '0%';
                progressBar.textContent = '0%';
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

    // Re-attach delete button listener if not already attached (in case it was added later)
    const attachDeleteButton = () => {
        const btn = document.getElementById('delete-course-btn');
        if (btn && !btn.hasAttribute('data-listener')) {
            btn.setAttribute('data-listener', 'true');
            btn.addEventListener('click', () => {
                const courseId = document.getElementById('course-id').value;
                if (courseId && courseId !== '0') {
                    openDeleteCourseModal(courseId);
                }
            });
        }
    };
    // Run now and also after load (e.g., when edit panel becomes visible)
    attachDeleteButton();
    // Observe if edit panel becomes visible (optional: call again when showCourseManage finishes)
    // We'll simply call it again inside showCourseManage after populating.
    // Override showCourseManage to attach after showing.
    const originalShow = showCourseManage;
    window.showCourseManage = async function(courseId) {
        await originalShow(courseId);
        attachDeleteButton();
    };
    // Also ensure that when edit panel is shown, we reattach.
    const observer = new MutationObserver(() => {
        const editPanel = document.getElementById('course-edit-panel');
        if (editPanel && !editPanel.classList.contains('hidden')) {
            attachDeleteButton();
        }
    });
    observer.observe(document.body, { attributes: true, subtree: true, attributeFilter: ['class'] });
}