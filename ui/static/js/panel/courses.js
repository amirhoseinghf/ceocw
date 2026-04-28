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
    document.getElementById('course-image').value = course.ImageUrl || '';
    document.getElementById('course-telegram').value = course.TelegramLink || '';
    document.getElementById('course-bale').value = course.BaleLink || '';
    document.getElementById('course-title-display').innerText = course.Title || 'بدون عنوان';
    
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

function initCourses() {
    document.getElementById('back-to-courses-list').addEventListener('click', () => {
        document.getElementById('courses-list').classList.remove('hidden');
        document.getElementById('course-edit-panel').classList.add('hidden');
        loadCourses();
    });
    document.getElementById('save-course-basic').addEventListener('click', async () => {
    const courseId = document.getElementById('course-id').value;
    if (!courseId || courseId === '0') {
        showToast('شناسه دوره معتبر نیست', false);
        return;
    }
    const basicData = {
        title: document.getElementById('course-title').value,
        shortName: document.getElementById('course-shortname').value,
        imageUrl: document.getElementById('course-image').value,
        telegramLink: document.getElementById('course-telegram').value,
        baleLink: document.getElementById('course-bale').value,
        teacherId: parseInt(document.getElementById('course-teacher').value),
        semesterId: parseInt(document.getElementById('course-semester').value)
    };
    // Simple validation
    if (!basicData.title || !basicData.shortName) {
        showToast('عنوان و نام کوتاه دوره الزامی است', false);
        return;
    }
    try {
        const response = await fetch(`/courses/${courseId}/basic`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(basicData)
        });
        if (!response.ok) throw new Error();
        showToast('اطلاعات پایه دوره ذخیره شد', true);
        // Update the displayed title in the header
        document.getElementById('course-title-display').innerText = basicData.title;
        // Optionally reload the course list later when returning to list
    } catch (err) {
        showToast('خطا در ذخیره اطلاعات پایه', false);
    }
});
}