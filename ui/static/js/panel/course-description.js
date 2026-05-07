// course-description.js
let currentGradeItems = [];
let currentScheduleItems = [];

const WEEK_DAYS = ['', 'شنبه', 'یکشنبه', 'دوشنبه', 'سه‌شنبه', 'چهارشنبه', 'پنجشنبه', 'جمعه'];

function normalizeGradeItem(item) {
    return {
        name: item?.name ?? item?.Name ?? '',
        percentage: item?.percentage ?? item?.Percentage ?? ''
    };
}

function normalizeScheduleItem(item) {
    return {
        day_of_week: item?.day_of_week ?? item?.DayOfWeek ?? '',
        start_time: item?.start_time ?? item?.StartTime ?? '',
        end_time: item?.end_time ?? item?.EndTime ?? '',
        location: item?.location ?? item?.Location ?? ''
    };
}

function syncGradeItemsFromDOM() {
    document.querySelectorAll('.grade-items-table tr[data-idx]').forEach(row => {
        const idx = Number(row.dataset.idx);
        if (!Number.isInteger(idx) || !currentGradeItems[idx]) return;
        currentGradeItems[idx].name = row.querySelector('.grade-name')?.value || '';
        currentGradeItems[idx].percentage = row.querySelector('.grade-percent')?.value || '';
    });
}

function syncScheduleItemsFromDOM() {
    document.querySelectorAll('.schedule-item-editor[data-idx]').forEach(row => {
        const idx = Number(row.dataset.idx);
        if (!Number.isInteger(idx) || !currentScheduleItems[idx]) return;
        currentScheduleItems[idx] = {
            day_of_week: row.querySelector('.schedule-day')?.value || '',
            start_time: row.querySelector('.schedule-start')?.value || '',
            end_time: row.querySelector('.schedule-end')?.value || '',
            location: row.querySelector('.schedule-location')?.value || ''
        };
    });
}

// ── Unsaved-changes tracking: description ──────────────────────────────────
let _descOriginal = null;

function snapshotDescription() {
    _descOriginal = document.getElementById('course-full-description')?.value ?? '';
    clearDescUnsaved();
}
function checkDescUnsaved() {
    if (_descOriginal === null) return;
    const current = document.getElementById('course-full-description')?.value ?? '';
    const col = document.getElementById('desc-description-col');
    if (col) col.classList.toggle('has-unsaved-changes', current !== _descOriginal);
}
function clearDescUnsaved() {
    document.getElementById('desc-description-col')?.classList.remove('has-unsaved-changes');
}

// ── Unsaved-changes tracking: class schedule ──────────────────────────────
let _scheduleOriginal = null;
let _gradeOriginal = null;

function scheduleSnapshotValue() {
    syncScheduleItemsFromDOM();
    return JSON.stringify(currentScheduleItems);
}

function snapshotSchedule() {
    _scheduleOriginal = scheduleSnapshotValue();
    clearScheduleUnsaved();
}
function checkScheduleUnsaved() {
    if (_scheduleOriginal === null) return;
    const col = document.getElementById('desc-schedule-col');
    if (col) col.classList.toggle('has-unsaved-changes', scheduleSnapshotValue() !== _scheduleOriginal);
}
function clearScheduleUnsaved() {
    document.getElementById('desc-schedule-col')?.classList.remove('has-unsaved-changes');
}

function gradeSnapshotValue() {
    syncGradeItemsFromDOM();
    return JSON.stringify(currentGradeItems);
}

function snapshotGrades() {
    _gradeOriginal = gradeSnapshotValue();
    clearGradeUnsaved();
}

function checkGradeUnsaved() {
    if (_gradeOriginal === null) return;
    document.getElementById('desc-grade-col')?.classList.toggle('has-unsaved-changes', gradeSnapshotValue() !== _gradeOriginal);
}

function clearGradeUnsaved() {
    document.getElementById('desc-grade-col')?.classList.remove('has-unsaved-changes');
}

// ── Schedule items render ─────────────────────────────────────────────────
function renderScheduleItems(items) {
    const container = document.getElementById('schedule-items-list');
    if (!container) return;
    if (!items.length) currentScheduleItems = [normalizeScheduleItem({})];

    container.innerHTML = currentScheduleItems.map((item, idx) => `
        <div class="schedule-item-editor" data-idx="${idx}">
            <div class="form-group">
                <label>روز هفته</label>
                <select class="form-control schedule-day">
                    ${WEEK_DAYS.map(day => `
                        <option value="${escapeHtml(day)}"${day === item.day_of_week ? ' selected' : ''}>${day || '— انتخاب کنید —'}</option>
                    `).join('')}
                </select>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>ساعت شروع</label>
                    <input type="text" class="form-control schedule-start" placeholder="14:00" value="${escapeHtml(item.start_time)}">
                </div>
                <div class="form-group">
                    <label>ساعت پایان</label>
                    <input type="text" class="form-control schedule-end" placeholder="16:00" value="${escapeHtml(item.end_time)}">
                </div>
            </div>
            <div class="schedule-item-footer">
                <div class="form-group">
                    <label>مکان</label>
                    <input type="text" class="form-control schedule-location" placeholder="سالن ۱" value="${escapeHtml(item.location)}">
                </div>
                <button type="button" class="btn-delete remove-schedule-item" data-idx="${idx}">حذف</button>
            </div>
        </div>
    `).join('');

    container.querySelectorAll('input, select').forEach(el => {
        el.addEventListener('input', checkScheduleUnsaved);
        el.addEventListener('change', checkScheduleUnsaved);
    });
    container.querySelectorAll('.remove-schedule-item').forEach(btn => {
        btn.addEventListener('click', () => {
            syncScheduleItemsFromDOM();
            currentScheduleItems.splice(Number(btn.dataset.idx), 1);
            if (!currentScheduleItems.length) currentScheduleItems.push(normalizeScheduleItem({}));
            renderScheduleItems(currentScheduleItems);
            checkScheduleUnsaved();
        });
    });
}

// ── Grade items render ─────────────────────────────────────────────────────
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
    document.querySelectorAll('.remove-grade-item').forEach(btn => {
        btn.addEventListener('click', () => {
            syncGradeItemsFromDOM();
            const idx = parseInt(btn.dataset.idx);
            currentGradeItems.splice(idx, 1);
            renderGradeItems(currentGradeItems);
            checkGradeUnsaved();
        });
    });
    document.querySelectorAll('.grade-name, .grade-percent').forEach(inp => {
        inp.addEventListener('input', checkGradeUnsaved);
        inp.addEventListener('change', checkGradeUnsaved);
    });
}

// ── Load (called when course management panel opens) ──────────────────────
function loadCourseDescription(course) {
    const desc = document.getElementById('course-full-description');
    if (desc) desc.value = course.CourseDescription?.Description || '';

    const scheduleItems = course.CourseDescription?.schedule_items || [];
    const fallbackSchedule = course.CourseDescription?.class_schedule;
    currentScheduleItems = scheduleItems.length
        ? scheduleItems.map(normalizeScheduleItem)
        : (fallbackSchedule?.day_of_week ? [normalizeScheduleItem(fallbackSchedule)] : [normalizeScheduleItem({})]);
    renderScheduleItems(currentScheduleItems);

    currentGradeItems = (course.CourseDescription?.grade_distribution || []).map(normalizeGradeItem);
    renderGradeItems(currentGradeItems);

    setTimeout(snapshotDescription, 0);
    setTimeout(snapshotSchedule, 0);
    setTimeout(snapshotGrades, 0);
}

// ── Init ──────────────────────────────────────────────────────────────────
function initCourseDescription() {
    const saveDescBtn = document.getElementById('save-description');
    if (saveDescBtn) {
        saveDescBtn.addEventListener('click', async () => {
            const courseId = document.getElementById('course-id').value;
            const description = document.getElementById('course-full-description').value;
            try {
                const res = await fetch(`/courses/${courseId}/description`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ Description: description })
                });
                if (!res.ok) throw new Error();
                snapshotDescription();
                showToast('توضیحات دوره ذخیره شد', true);
            } catch {
                showToast('خطا در ذخیره توضیحات', false);
            }
        });
    }

    const saveScheduleBtn = document.getElementById('save-schedule');
    if (saveScheduleBtn) {
        saveScheduleBtn.addEventListener('click', async () => {
            const courseId = document.getElementById('course-id').value;
            syncScheduleItemsFromDOM();
            const items = currentScheduleItems.filter(item =>
                item.day_of_week || item.start_time || item.end_time || item.location);
            try {
                const res = await fetch(`/courses/${courseId}/schedule`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ schedule_items: items })
                });
                if (!res.ok) throw new Error();
                currentScheduleItems = items.length ? items : [normalizeScheduleItem({})];
                renderScheduleItems(currentScheduleItems);
                snapshotSchedule();
                showToast('برنامه کلاسی ذخیره شد', true);
            } catch {
                showToast('خطا در ذخیره برنامه', false);
            }
        });
    }

    const saveGradeBtn = document.getElementById('save-grade-distribution');
    if (saveGradeBtn) {
        saveGradeBtn.addEventListener('click', async () => {
            const courseId = document.getElementById('course-id').value;
            syncGradeItemsFromDOM();
            const items = currentGradeItems.filter(item => item.name || item.percentage);
            try {
                const res = await fetch(`/courses/${courseId}/grade-items`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(items)
                });
                if (!res.ok) throw new Error();
                currentGradeItems = items;
                renderGradeItems(currentGradeItems);
                snapshotGrades();
                showToast('توزیع نمرات ذخیره شد', true);
            } catch {
                showToast('خطا در ذخیره توزیع نمرات', false);
            }
        });
    }

    const addItemBtn = document.getElementById('add-grade-item');
    if (addItemBtn) {
        addItemBtn.addEventListener('click', () => {
            syncGradeItemsFromDOM();
            currentGradeItems.push({ name: '', percentage: '' });
            renderGradeItems(currentGradeItems);
            checkGradeUnsaved();
        });
    }

    const addScheduleBtn = document.getElementById('add-schedule-item');
    if (addScheduleBtn) {
        addScheduleBtn.addEventListener('click', () => {
            syncScheduleItemsFromDOM();
            currentScheduleItems.push(normalizeScheduleItem({}));
            renderScheduleItems(currentScheduleItems);
            checkScheduleUnsaved();
        });
    }

    const descArea = document.getElementById('course-full-description');
    if (descArea) descArea.addEventListener('input', checkDescUnsaved);
}
