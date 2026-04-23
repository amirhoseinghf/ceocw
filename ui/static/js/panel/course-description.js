// course-description.js
let currentGradeItems = [];

function renderGradeItems(items) {
    const container = document.getElementById('grade-items-list');
    if (!container) return;
    if (!items.length) {
        container.innerHTML = '<p class="loading">هیچ آیتمی تعریف نشده است.</p>';
        return;
    }
    const html = `
        <div class="grade-items-container">
            <table class="teachers-table grade-items-table">
                <thead>
                    <tr><th>نام</th><th>درصد</th><th>عملیات</th></tr>
                </thead>
                <tbody>
                    ${items.map((item, idx) => `
                        <tr data-idx="${idx}">
                            <td data-label="نام"><input type="text" class="grade-name" data-idx="${idx}" value="${escapeHtml(item.name)}" style="width:100%;"></td>
                            <td data-label="درصد"><input type="text" class="grade-percent" data-idx="${idx}" value="${escapeHtml(item.percentage)}" style="width:100%;"></td>
                            <td data-label="عملیات"><button class="btn-delete remove-grade-item" data-idx="${idx}">حذف</button></td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        </div>
    `;
    container.innerHTML = html;
    // attach remove events
    document.querySelectorAll('.remove-grade-item').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const idx = parseInt(btn.dataset.idx);
            currentGradeItems.splice(idx, 1);
            renderGradeItems(currentGradeItems);
        });
    });
    // attach change events to update currentGradeItems
    document.querySelectorAll('.grade-name').forEach(inp => {
        inp.addEventListener('change', (e) => {
            const idx = parseInt(inp.dataset.idx);
            currentGradeItems[idx].name = inp.value;
        });
    });
    document.querySelectorAll('.grade-percent').forEach(inp => {
        inp.addEventListener('change', (e) => {
            const idx = parseInt(inp.dataset.idx);
            currentGradeItems[idx].percentage = inp.value;
        });
    });
}

function loadCourseDescription(course) {
    // Populate description textarea
    document.getElementById('course-full-description').value = course.CourseDescription?.Description || '';
    // Populate class schedule
    document.getElementById('schedule-day').value = course.CourseDescription?.class_schedule?.day_of_week || '';
    document.getElementById('schedule-start').value = course.CourseDescription?.class_schedule?.start_time || '';
    document.getElementById('schedule-end').value = course.CourseDescription?.class_schedule?.end_time || '';
    document.getElementById('schedule-location').value = course.CourseDescription?.class_schedule?.location || '';
    // Populate grade distribution
    currentGradeItems = course.CourseDescription?.grade_distribution || [];
    renderGradeItems(currentGradeItems);
}

function initCourseDescription() {
    // Save only the description text
    const saveDescBtn = document.getElementById('save-description');
    if (saveDescBtn) {
        saveDescBtn.addEventListener('click', async () => {
            const courseId = document.getElementById('course-id').value;
            const description = document.getElementById('course-full-description').value;
            try {
                const response = await fetch(`/courses/${courseId}/description`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ Description: description })
                });
                if (!response.ok) throw new Error();
                showToast('توضیحات دوره ذخیره شد', true);
            } catch (err) {
                showToast('خطا در ذخیره توضیحات', false);
            }
        });
    }

    // Save class schedule (no description)
    const saveScheduleBtn = document.getElementById('save-schedule');
    if (saveScheduleBtn) {
        saveScheduleBtn.addEventListener('click', async () => {
            const courseId = document.getElementById('course-id').value;
            const data = {
                ClassSchedule: {
                    DayOfWeek: document.getElementById('schedule-day').value,
                    StartTime: document.getElementById('schedule-start').value,
                    EndTime: document.getElementById('schedule-end').value,
                    Location: document.getElementById('schedule-location').value
                }
            };
            try {
                const response = await fetch(`/courses/${courseId}/schedule`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                if (!response.ok) throw new Error();
                showToast('برنامه کلاسی ذخیره شد', true);
            } catch (err) {
                showToast('خطا در ذخیره برنامه', false);
            }
        });
    }

    // Save grade distribution (unchanged)
    const saveGradeBtn = document.getElementById('save-grade-distribution');
    if (saveGradeBtn) {
        saveGradeBtn.addEventListener('click', async () => {
            const courseId = document.getElementById('course-id').value;
            try {
                const response = await fetch(`/courses/${courseId}/grade-items`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(currentGradeItems)
                });
                if (!response.ok) throw new Error();
                showToast('توزیع نمرات ذخیره شد', true);
            } catch (err) {
                showToast('خطا در ذخیره توزیع نمرات', false);
            }
        });
    }

    // Add new grade item (unchanged)
    const addItemBtn = document.getElementById('add-grade-item');
    if (addItemBtn) {
        addItemBtn.addEventListener('click', () => {
            currentGradeItems.push({ name: '', percentage: '' });
            renderGradeItems(currentGradeItems);
        });
    }
}

// Make sure to call this from main.js after DOM ready
// window.initCourseDescription = initCourseDescription;