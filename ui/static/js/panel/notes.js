// notes.js
async function loadNotes(courseId) {
    const container = document.getElementById('notes-list');
    if (!container) return;
    container.innerHTML = '<div class="loading">در حال بارگذاری...</div>';
    try {
        const response = await fetch(`/courses/${courseId}/notes`);
        if (!response.ok) throw new Error();
        const notes = await response.json();
        renderNotes(notes);
    } catch (err) {
        console.error(err);
        container.innerHTML = '<div class="loading">خطا در بارگذاری جزوه‌ها</div>';
    }
}

function renderNotes(notes) {
        
    const container = document.getElementById('notes-list');
    if (!container) return;
    if (!Array.isArray(notes)) {
        container.innerHTML = '<div class="loading">خطا در دریافت اطلاعات</div>';
        return;
    }
    if (notes.length === 0) {
        container.innerHTML = '<div class="loading">هیچ جزوه‌ای ثبت نشده است.</div>';
        return;
    }
    const html = `
        <table class="teachers-table">
            <thead>
                <tr>
                    <th>عنوان</th>
                    <th>بروزرسانی</th>
                    <th>فایل</th>
                    <th>عملیات</th>
                </tr>
            </thead>
            <tbody>
                ${notes.map(note => `
                    <tr data-id="${note.Id}">
                        <td>${escapeHtml(note.Title)}</td>
                        <td>${note.IsUpdated ? '✅ بله' : '❌ خیر'}</td>
                        <td><a href="${escapeHtml(note.FileName)}" target="_blank">دانلود</a></td>
                        <td class="teacher-actions">
                        <button class="btn btn-edit edit-note" data-id="${note.Id}">✏️ ویرایش</button>
                            <button class="btn btn-delete delete-note" data-id="${note.Id}">🗑️ حذف</button>
                         </td>
                     </tr>
                `).join('')}
            </tbody>
         </table>
    `;
    container.innerHTML = html;
    document.querySelectorAll('.delete-note').forEach(btn => {
        btn.addEventListener('click', async () => {
            if (confirm('آیا از حذف این جزوه اطمینان دارید؟')) {
                await deleteNote(parseInt(btn.dataset.id));
                const courseId = document.getElementById('course-id').value;
                if (courseId) loadNotes(courseId);
            }
        });
    });
    document.querySelectorAll('.edit-note').forEach(btn => {
        btn.addEventListener('click', () => openNoteModal(parseInt(btn.dataset.id)));
    });
}

async function deleteNote(noteId) {
    const response = await fetch(`/notes/${noteId}`, { method: 'DELETE' });
    if (response.ok) {
        showToast('جزوه حذف شد', true);
        return true;
    } else {
        showToast('خطا در حذف جزوه', false);
        return false;
    }
}

async function openNoteModal(noteId = 0) {
    console.log(noteId)
    const modal = document.getElementById('note-modal');
    const form = document.getElementById('note-form');
    form.reset();
    document.getElementById('note-id').value = '0';
    document.getElementById('note-file').value = '';
    document.getElementById('note-is-updated').checked = false;

    
    if (noteId) {
        try {
            const response = await fetch(`/notes/${noteId}`);
            const note = await response.json();
            document.getElementById('note-id').value = note.Id;
            document.getElementById('note-title').value = note.Title;
            document.getElementById('note-is-updated').checked = note.IsUpdated;
            document.getElementById('note-modal-title').innerText = 'ویرایش جزوه';
            // Note: file input remains empty; user can upload new file to replace, otherwise keep old.
        } catch (err) {
            showToast('خطا در دریافت اطلاعات جزوه', false);
            return;
        }
    } else {
        document.getElementById('note-modal-title').innerText = 'افزودن جزوه جدید';
    }
    modal.style.display = 'flex';
}

function closeNoteModal() {
    document.getElementById('note-modal').style.display = 'none';
}

function initNotes() {
    const addBtn = document.getElementById('add-note-btn');
    if (addBtn) {
        addBtn.addEventListener('click', () => openNoteModal(0));
    } 
    document.querySelector('.note-close')?.addEventListener('click', closeNoteModal);
    document.getElementById('note-cancel')?.addEventListener('click', closeNoteModal);
    window.addEventListener('click', (e) => {
        if (e.target === document.getElementById('note-modal')) closeNoteModal();
    });

    const form = document.getElementById('note-form');
form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const courseId = document.getElementById('course-id').value;
    const noteId = document.getElementById('note-id').value;
    const title = document.getElementById('note-title').value;
    const isUpdated = document.getElementById('note-is-updated').checked;
    const fileInput = document.getElementById('note-file');
    const file = fileInput.files[0];

    if (!title) {
        showToast('عنوان جزوه الزامی است', false);
        return;
    }
    // For new note, file is required. For update, file is optional (keep old if not provided)
    if (!noteId && !file) {
        showToast('فایل جزوه الزامی است', false);
        return;
    }

    const formData = new FormData();
    formData.append('title', title);
    formData.append('is_updated', isUpdated);
    formData.append('course_id', courseId); // needed for file saving during update
    if (file) formData.append('note_file', file);

    let url, method;
    if (noteId && noteId !== '0') {
        url = `/notes/${noteId}`;
        method = 'PUT';
        formData.append('id', noteId);
    } else {
        url = `/courses/${courseId}/notes`;
        method = 'POST';
    }

    // Progress bar only if file is uploaded
    const progressContainer = document.getElementById('note-progress-container');
    const progressBar = document.getElementById('note-progress-bar');
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
            showToast(noteId ? 'جزوه ویرایش شد' : 'جزوه اضافه شد', true);
            closeNoteModal();
            loadNotes(courseId);
        } else {
            showToast('خطا در ذخیره جزوه', false);
        }
    };
    xhr.onerror = () => {
        if (progressContainer) progressContainer.style.display = 'none';
        showToast('خطا در شبکه', false);
    };
    xhr.send(formData);
});
}