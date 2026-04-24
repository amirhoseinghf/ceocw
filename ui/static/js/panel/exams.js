// exams.js
async function loadExams(courseId) {
    const container = document.getElementById('exams-list');
    if (!container) return;
    container.innerHTML = '<div class="loading">در حال بارگذاری...</div>';
    try {
        const response = await fetch(`/courses/${courseId}/exams`);
        if (!response.ok) throw new Error();
        const exams = await response.json();
        renderExams(exams);
    } catch (err) {
        console.error(err);
        container.innerHTML = '<div class="loading">خطا در بارگذاری امتحانات</div>';
    }
}

function renderExams(exams) {
    const container = document.getElementById('exams-list');
    if (!container) return;
    if (!Array.isArray(exams)) {
        container.innerHTML = '<div class="loading">خطا در دریافت اطلاعات</div>';
        return;
    }
    if (exams.length === 0) {
        container.innerHTML = '<div class="loading">هیچ امتحانی ثبت نشده است.</div>';
        return;
    }
    
    const html = `
        <table class="teachers-table">
            <thead>
                <tr>
                    <th>ترم</th>
                    <th>نوع امتحان</th>
                    <th>این ترم</th>
                    <th>فایل</th>
                    <th>عملیات</th>
                </tr>
            </thead>
            <tbody>
                ${exams.map(exam => {
                    // Compute semester name manually
                    let semesterName = 'نامشخص';
                    if (exam.Semester && exam.Semester.Year) {
                        const seasonPersian = exam.Semester.Season === 'spring' ? 'بهار' : 'پاییز';
                        semesterName = `${seasonPersian} ${exam.Semester.Year}`;
                    }
                    const examTypePersian = exam.ExamType === 'Midterm' ? 'میان‌ترم' : (exam.ExamType === 'Final' ? 'پایان‌ترم' : 'کوییز');
                    const thisSemesterText = exam.ThisSemester ? '✅ بله' : '❌ خیر';
                    return `
                        <tr data-id="${exam.Id}">
                            <td>${semesterName}</td>
                            <td>${examTypePersian}</td>
                            <td>${thisSemesterText}</td>
                            <td><a href="${escapeHtml(exam.FileName)}" target="_blank">دانلود</a></td>
                            <td class="teacher-actions">
                                <button class="btn btn-edit edit-exam" data-id="${exam.Id}">✏️ ویرایش</button>
                                <button class="btn btn-delete delete-exam" data-id="${exam.Id}">🗑️ حذف</button>
                            </td>
                        </table>
                    `;
                }).join('')}
            </tbody>
        </table>
    `;
    container.innerHTML = html;
    document.querySelectorAll('.edit-exam').forEach(btn => {
        btn.addEventListener('click', () => openExamModal(parseInt(btn.dataset.id)));
    });
    document.querySelectorAll('.delete-exam').forEach(btn => {
        btn.addEventListener('click', async () => {
            if (confirm('آیا از حذف این امتحان اطمینان دارید؟')) {
                await deleteExam(parseInt(btn.dataset.id));
                const courseId = document.getElementById('course-id').value;
                if (courseId) loadExams(courseId);
            }
        });
    });
}

async function deleteExam(examId) {
    const response = await fetch(`/exams/${examId}`, { method: 'DELETE' });
    if (response.ok) {
        showToast('امتحان حذف شد', true);
        return true;
    } else {
        showToast('خطا در حذف امتحان', false);
        return false;
    }
}

async function openExamModal(examId = 0) {
    // Load semesters into dropdown if not already populated
    const semesterSelect = document.getElementById('exam-semester');
    if (semesterSelect.options.length === 0) {
        await loadSemesterOptionsForExams();
    }
    const modal = document.getElementById('exam-modal');
    const form = document.getElementById('exam-form');
    form.reset();
    document.getElementById('exam-id').value = '0';
    document.getElementById('exam-file').value = '';
    document.getElementById('exam-this-semester').checked = false;

    if (examId) {
        try {
            const response = await fetch(`/exams/${examId}`);
            const exam = await response.json();
            document.getElementById('exam-id').value = exam.Id;
            document.getElementById('exam-type').value = exam.ExamType;
            document.getElementById('exam-this-semester').checked = exam.ThisSemester;
            if (exam.Semester && exam.Semester.Id) {
                document.getElementById('exam-semester').value = exam.Semester.Id;
            }
            document.getElementById('exam-modal-title').innerText = 'ویرایش امتحان';
        } catch (err) {
            showToast('خطا در دریافت اطلاعات امتحان', false);
            return;
        }
    } else {
        document.getElementById('exam-modal-title').innerText = 'افزودن امتحان جدید';
    }
    modal.style.display = 'flex';
}

async function loadSemesterOptionsForExams() {
    try {
        const response = await fetch('/semesters');
        const semesters = await response.json();
        const select = document.getElementById('exam-semester');
        select.innerHTML = '<option value="">انتخاب ترم</option>' +
            semesters.map(s => `<option value="${s.Id}">${s.Season === 'spring' ? 'بهار' : 'پاییز'} ${s.Year}</option>`).join('');
    } catch (err) {
        console.error('Error loading semesters:', err);
    }
}

function closeExamModal() {
    document.getElementById('exam-modal').style.display = 'none';
}

function initExams() {
    // Preload semester options
    loadSemesterOptionsForExams();

    const addBtn = document.getElementById('add-exam-btn');
    if (addBtn) addBtn.addEventListener('click', () => openExamModal(0));
    document.querySelector('.exam-close')?.addEventListener('click', closeExamModal);
    document.getElementById('exam-cancel')?.addEventListener('click', closeExamModal);
    window.addEventListener('click', (e) => {
        if (e.target === document.getElementById('exam-modal')) closeExamModal();
    });

    const form = document.getElementById('exam-form');
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const courseId = document.getElementById('course-id').value;
        const examId = document.getElementById('exam-id').value;
        const semesterId = document.getElementById('exam-semester').value;
        const examType = document.getElementById('exam-type').value;
        const thisSemester = document.getElementById('exam-this-semester').checked;
        const fileInput = document.getElementById('exam-file');
        const file = fileInput.files[0];

        
        if (!examType) {
            showToast('نوع امتحان الزامی است', false);
            return;
        }
        if (examId == 0 && !file) {
            showToast('فایل امتحان الزامی است', false);
            return;
        }

        const formData = new FormData();
        formData.append('exam_type', examType);
        formData.append('this_semester', thisSemester);
        if (semesterId) formData.append('semester_id', semesterId);
        if (file) formData.append('exam_file', file);
        formData.append('course_id', courseId); // for file saving during update

        let url, method;
        if (examId && examId !== '0') {
            url = `/exams/${examId}`;
            method = 'PUT';
            formData.append('id', examId);
        } else {
            url = `/courses/${courseId}/exams`;
            method = 'POST';
        }

        const progressContainer = document.getElementById('exam-progress-container');
        const progressBar = document.getElementById('exam-progress-bar');
        if (file) {
            progressContainer.style.display = 'block';
            progressBar.style.width = '0%';
            progressBar.textContent = '0%';
        }

        const xhr = new XMLHttpRequest();
        xhr.open(method, url, true);
        if (file) {
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
                showToast(examId ? 'امتحان ویرایش شد' : 'امتحان اضافه شد', true);
                closeExamModal();
                loadExams(courseId);
            } else {
                showToast('خطا در ذخیره امتحان', false);
            }
        };
        xhr.onerror = () => {
            if (progressContainer) progressContainer.style.display = 'none';
            showToast('خطا در شبکه', false);
        };
        xhr.send(formData);
    });
}