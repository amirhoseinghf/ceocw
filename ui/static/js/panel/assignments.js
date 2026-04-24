// assignments.js
async function loadAssignments(courseId) {
    const container = document.getElementById('assignments-list');
    if (!container) return;
    container.innerHTML = '<div class="loading">در حال بارگذاری...</div>';
    try {
        const response = await fetch(`/courses/${courseId}/assignments`);
        if (!response.ok) throw new Error();
        const assignments = await response.json();
        renderAssignments(assignments);
    } catch (err) {
        console.error(err);
        container.innerHTML = '<div class="loading">خطا در بارگذاری تکالیف</div>';
    }
}

function renderAssignments(assignments) {
    const container = document.getElementById('assignments-list');
    if (!container) return;
    if (!Array.isArray(assignments)) {
        container.innerHTML = '<div class="loading">خطا در دریافت اطلاعات</div>';
        return;
    }
    if (assignments.length === 0) {
        container.innerHTML = '<div class="loading">هیچ تکلیفی ثبت نشده است.</div>';
        return;
    }
    const html = `
        <table class="teachers-table">
            <thead>
                <tr>
                    <th>عنوان</th>
                    <th>توضیحات</th>
                    <th>تاریخ انتشار</th>
                    <th>ددلاین</th>
                    <th>فایل</th>
                    <th>حل</th>
                    <th>نوع</th>
                    <th>تمدید شده</th>
                    <th>عملیات</th>
                </tr>
            </thead>
            <tbody>
                ${assignments.map(a => `
                    <tr data-id="${a.Id}">
                        <td>${escapeHtml(a.Title)}</td>
                        <td class="assignment-description" title="${escapeHtml(a.Description || '')}">${escapeHtml(a.Description || '—')}</td>
                        <td>${a.ReleaseDate ? new Date(a.ReleaseDate).toLocaleDateString('fa-IR') : '—'}</td>
                        <td>${a.DeadlineDate ? new Date(a.DeadlineDate).toLocaleDateString('fa-IR') : '—'}</td>
                        <td>${a.FileName ? `<a href="${escapeHtml(a.FileName)}" target="_blank">دانلود</a>` : '—'}</td>
                        <td>${a.SolutionName ? `<a href="${escapeHtml(a.SolutionName)}" target="_blank">دانلود</a>` : '—'}</td>
                        <td>${a.IsProject ? 'پروژه' : 'تکلیف'}</td>
                        <td>${a.IsExtended ? '✅ بله' : '❌ خیر'}</td>
                        <td class="teacher-actions">
                            <button class="btn btn-edit edit-assignment" data-id="${a.Id}">✏️ ویرایش</button>
                            <button class="btn btn-delete delete-assignment" data-id="${a.Id}">🗑️ حذف</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
    container.innerHTML = html;
    // Re-attach edit and delete events
    document.querySelectorAll('.edit-assignment').forEach(btn => {
        btn.addEventListener('click', () => openAssignmentModal(parseInt(btn.dataset.id)));
    });
    document.querySelectorAll('.delete-assignment').forEach(btn => {
        btn.addEventListener('click', async () => {
            if (confirm('آیا از حذف این تکلیف اطمینان دارید؟')) {
                await deleteAssignment(parseInt(btn.dataset.id));
                const courseId = document.getElementById('course-id').value;
                if (courseId) loadAssignments(courseId);
            }
        });
    });
}

async function deleteAssignment(assignmentId) {
    const response = await fetch(`/assignments/${assignmentId}`, { method: 'DELETE' });
    if (response.ok) {
        showToast('تکلیف حذف شد', true);
        return true;
    } else {
        showToast('خطا در حذف تکلیف', false);
        return false;
    }
}

async function openAssignmentModal(assignmentId = 0) {
    const modal = document.getElementById('assignment-modal');
    const form = document.getElementById('assignment-form');
    form.reset();
    document.getElementById('assignment-id').value = '0';
    document.getElementById('assignment-file').value = '';
    document.getElementById('assignment-solution').value = '';
    document.getElementById('assignment-is-extended').checked = false;
    document.getElementById('assignment-is-project').checked = false;

    if (assignmentId) {
        try {
            const response = await fetch(`/assignments/${assignmentId}`);
            const assignment = await response.json();
            document.getElementById('assignment-id').value = assignment.Id;
            document.getElementById('assignment-title').value = assignment.Title;
            document.getElementById('assignment-description').value = assignment.Description || '';
            if (assignment.ReleaseDate) {
                const release = new Date(assignment.ReleaseDate);
                document.getElementById('assignment-release-date').value = release.toISOString().slice(0, 16);
            }
            if (assignment.DeadlineDate) {
                const deadline = new Date(assignment.DeadlineDate);
                document.getElementById('assignment-deadline-date').value = deadline.toISOString().slice(0, 16);
            }
            document.getElementById('assignment-is-extended').checked = assignment.IsExtended;
            document.getElementById('assignment-is-project').checked = assignment.IsProject;
            document.getElementById('assignment-modal-title').innerText = 'ویرایش تکلیف';
        } catch (err) {
            showToast('خطا در دریافت اطلاعات تکلیف', false);
            return;
        }
    } else {
        document.getElementById('assignment-modal-title').innerText = 'افزودن تکلیف جدید';
    }
    modal.style.display = 'flex';
}

function closeAssignmentModal() {
    document.getElementById('assignment-modal').style.display = 'none';
}

function initAssignments() {
    const addBtn = document.getElementById('add-assignment-btn');
    if (addBtn) addBtn.addEventListener('click', () => openAssignmentModal(0));
    document.querySelector('.assignment-close')?.addEventListener('click', closeAssignmentModal);
    document.getElementById('assignment-cancel')?.addEventListener('click', closeAssignmentModal);
    window.addEventListener('click', (e) => {
        if (e.target === document.getElementById('assignment-modal')) closeAssignmentModal();
    });

    const form = document.getElementById('assignment-form');
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const courseId = document.getElementById('course-id').value;
        const assignmentId = document.getElementById('assignment-id').value;
        const title = document.getElementById('assignment-title').value;
        const description = document.getElementById('assignment-description').value;
        const releaseDate = document.getElementById('assignment-release-date').value;
        const deadlineDate = document.getElementById('assignment-deadline-date').value;
        const isExtended = document.getElementById('assignment-is-extended').checked;
        const isProject = document.getElementById('assignment-is-project').checked;
        const assignmentFile = document.getElementById('assignment-file').files[0];
        const solutionFile = document.getElementById('assignment-solution').files[0];

        const formData = new FormData();
        formData.append('title', title);
        formData.append('description', description);
        if (releaseDate) formData.append('release_date', releaseDate);
        if (deadlineDate) formData.append('deadline_date', deadlineDate);
        formData.append('is_extended', isExtended);
        formData.append('is_project', isProject);
        formData.append('course_id', courseId);
        if (assignmentFile) formData.append('assignment_file', assignmentFile);
        if (solutionFile) formData.append('solution_file', solutionFile);

        let url, method;
        if (assignmentId && assignmentId !== '0') {
            url = `/assignments/${assignmentId}`;
            method = 'PUT';
            formData.append('id', assignmentId);
        } else {
            url = `/courses/${courseId}/assignments`;
            method = 'POST';
        }

        const progressContainer = document.getElementById('assignment-progress-container');
        const progressBar = document.getElementById('assignment-progress-bar');
        if (assignmentFile) {
            progressContainer.style.display = 'block';
            progressBar.style.width = '0%';
            progressBar.textContent = '0%';
        }

        const xhr = new XMLHttpRequest();
        xhr.open(method, url, true);
        if (assignmentFile) {
            xhr.upload.addEventListener('progress', (ev) => {
                if (ev.lengthComputable) {
                    const percent = (ev.loaded / ev.total) * 100;
                    progressBar.style.width = percent + '%';
                    progressBar.textContent = Math.round(percent) + '%';
                }
            });
        }
        xhr.onload = () => {
            progressContainer.style.display = 'none';
            if (xhr.status === 200 || xhr.status === 201) {
                showToast(assignmentId ? 'تکلیف ویرایش شد' : 'تکلیف اضافه شد', true);
                closeAssignmentModal();
                loadAssignments(courseId);
            } else {
                showToast('خطا در ذخیره تکلیف', false);
            }
        };
        xhr.onerror = () => {
            progressContainer.style.display = 'none';
            showToast('خطا در شبکه', false);
        };
        xhr.send(formData);
    });
}