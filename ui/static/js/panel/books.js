// books.js
let currentBookId = 0;

let pendingDeleteBookId = null;
let pendingDeleteCourseId = null;

async function loadBooks(courseId) {
    const container = document.getElementById('books-list');
    container.innerHTML = '<div class="loading">در حال بارگذاری...</div>';
    try {
        const response = await fetch(`/courses/${courseId}/books`);
        if (!response.ok) throw new Error();
        const books = await response.json();
        renderBooks(books);
    } catch (err) {
        container.innerHTML = '<div class="loading">خطا در بارگذاری کتاب‌ها</div>';
        console.error(err);
    }
}

function renderBooks(books) {
    const container = document.getElementById('books-list');
    if (!container) return;
    
    // Add a cache-busting timestamp for this render pass
    const cacheBuster = Date.now();

    if (!Array.isArray(books)) {
        container.innerHTML = '<div class="loading">خطا در دریافت اطلاعات کتاب‌ها</div>';
        return;
    }
    if (!books.length) {
        container.innerHTML = '<div class="loading">هیچ کتابی ثبت نشده است.</div>';
        return;
    }
    
    const html = `
        <table class="teachers-table">
            <thead>
                <tr><th>عنوان</th><th>جلد</th><th>فایل / لینک</th><th>عملیات</th></tr>
            </thead>
            <tbody>
                ${books.map(book => `
                    <tr data-id="${book.Id}">
                        <td>${escapeHtml(book.Title)}</td>
                        <td>
                            <img class="book-thumbnail" 
                                 src="${book.ImageURL ? book.ImageURL + '?_=' + cacheBuster : 'https://via.placeholder.com/50'}" 
                                 onerror="this.src='https://via.placeholder.com/50'">
                         </td>
                        <td>${book.DownloadURL ? `<a href="${escapeHtml(book.DownloadURL)}" target="_blank">دانلود</a>` : '—'}</td>
                        <td class="teacher-actions">
                            <button class="btn btn-edit edit-book" data-id="${book.Id}">✏️ ویرایش</button>
                            <button class="btn btn-delete delete-book" data-id="${book.Id}">🗑️ حذف</button>
                        </td>
                     </tr>
                `).join('')}
            </tbody>
        </table>
    `;
    container.innerHTML = html;
    
    // Re-attach event listeners
    document.querySelectorAll('.edit-book').forEach(btn => {
        btn.addEventListener('click', () => openBookModal(parseInt(btn.dataset.id)));
    });
    document.querySelectorAll('.delete-book').forEach(btn => {
        btn.addEventListener('click', () => {
            const bookId = parseInt(btn.dataset.id);
            const courseId = parseInt(document.getElementById('course-id').value);
            showBookDeleteOptions(bookId, courseId);
        });
    });
}

function showBookDeleteOptions(bookId, courseId) {
    pendingDeleteBookId = bookId;
    pendingDeleteCourseId = courseId;
    const modal = document.getElementById('delete-book-modal');
    modal.style.display = 'flex';
}

function closeBookDeleteModal() {
    document.getElementById('delete-book-modal').style.display = 'none';
    pendingDeleteBookId = null;
}

async function detachBookFromCourse() {
    if (!pendingDeleteBookId || !pendingDeleteCourseId) return;
    try {
        const response = await fetch(`/courses/${pendingDeleteCourseId}/books/${pendingDeleteBookId}`, {
            method: 'DELETE'
        });
        if (!response.ok) throw new Error();
        showToast('کتاب از این دوره حذف شد', true);
        closeBookDeleteModal();
        // Refresh the book list for the current course
        loadBooks(pendingDeleteCourseId);
    } catch (err) {
        showToast('خطا در حذف کتاب از دوره', false);
        closeBookDeleteModal();
    }
}

async function deleteBookPermanently() {
    if (!pendingDeleteBookId) return;
    try {
        const response = await fetch(`/books/${pendingDeleteBookId}`, {
            method: 'DELETE'
        });
        if (!response.ok) throw new Error();
        showToast('کتاب به طور کامل حذف شد', true);
        closeBookDeleteModal();
        // Refresh the book list for the current course (the book is gone)
        if (pendingDeleteCourseId) console.log("SFDSF")
        if (pendingDeleteCourseId) loadBooks(pendingDeleteCourseId);
    } catch (err) {
        showToast('خطا در حذف کامل کتاب', false);
        closeBookDeleteModal();
    }
}

async function openBookModal(bookId = 0) {
    currentBookId = bookId;
    const modal = document.getElementById('book-modal');
    const form = document.getElementById('book-form');
    form.reset();
    document.getElementById('book-id').value = '0';
    document.getElementById('book-file').value = '';
    document.getElementById('book-thumbnail').value = '';
    document.getElementById('book-download-url').value = '';

    // Hide previews by default
    const thumbCurrent = document.getElementById('book-thumbnail-current');
    const fileCurrent  = document.getElementById('book-file-current');
    if (thumbCurrent) thumbCurrent.style.display = 'none';
    if (fileCurrent)  fileCurrent.style.display  = 'none';

    if (bookId) {
        try {
            const response = await fetch(`/books/${bookId}`);
            const book = await response.json();
            document.getElementById('book-id').value = book.Id;
            document.getElementById('book-title').value = book.Title;

            // Show existing thumbnail if any
            if (book.ImageURL && thumbCurrent) {
                document.getElementById('book-thumbnail-img').src = book.ImageURL + '?_=' + Date.now();
                thumbCurrent.style.display = 'flex';
            }

            // Show existing file/link
            const downloadUrl = book.DownloadURL || '';
            if (downloadUrl) {
                if (downloadUrl.startsWith('/data/')) {
                    // Local file: show indicator
                    if (fileCurrent) {
                        document.getElementById('book-file-current-link').href = downloadUrl;
                        fileCurrent.style.display = 'flex';
                    }
                    document.getElementById('book-download-url').value = '';
                } else {
                    // External link: keep it editable in the URL field
                    document.getElementById('book-download-url').value = downloadUrl;
                }
            }

            document.getElementById('book-modal-title').innerText = 'ویرایش کتاب';
        } catch (err) {
            showToast('خطا در دریافت اطلاعات کتاب', false);
            return;
        }
    } else {
        document.getElementById('book-modal-title').innerText = 'افزودن کتاب جدید';
    }
    modal.style.display = 'flex';

    // Hide progress bars and reset values
    const thumbContainer = document.getElementById('thumb-progress-container');
    const pdfContainer = document.getElementById('pdf-progress-container');
    if (thumbContainer) thumbContainer.style.display = 'none';
    if (pdfContainer) pdfContainer.style.display = 'none';
}


// Function to open the existing books modal
async function openExistingBooksModal() {
    const modal = document.getElementById('existing-books-modal');
    const container = document.getElementById('existing-books-list');
    modal.style.display = 'flex';
    await loadAllBooks();
}

async function loadAllBooks(searchTerm = '') {
    const container = document.getElementById('existing-books-list');
    const courseId = document.getElementById('course-id').value;
    container.innerHTML = '<div class="loading">در حال بارگذاری...</div>';
    try {
        let url = '/books';
        if (searchTerm) url += `?search=${encodeURIComponent(searchTerm)}`;
        const response = await fetch(url);
        if (!response.ok) throw new Error();
        const books = await response.json();
        // Also fetch currently attached books to grey them out or show badge
        const attachedResponse = await fetch(`/courses/${courseId}/books`);
        const attachedBooks = await attachedResponse.json();
        const attachedIds = new Set(attachedBooks.map(b => b.Id));
        renderExistingBooks(books, attachedIds);
    } catch (err) {
        container.innerHTML = '<div class="loading">خطا در بارگذاری کتاب‌ها</div>';
        console.error(err);
    }
}

function renderExistingBooks(books, attachedIds) {
    const container = document.getElementById('existing-books-list');
    if (!books.length) {
        container.innerHTML = '<div class="loading">کتابی یافت نشد.</div>';
        return;
    }
    const html = books.map(book => `
        <div class="book-item" data-id="${book.Id}" data-title="${escapeHtml(book.Title)}">
            <img src="${book.ImageURL || 'https://via.placeholder.com/50'}" onerror="this.src='https://via.placeholder.com/50'">
            <div class="book-title">${escapeHtml(book.Title)}</div>
            ${attachedIds.has(book.Id) ? '<div class="book-attached">✅ قبلاً اضافه شده</div>' : ''}
        </div>
    `).join('');
    container.innerHTML = html;
    // Attach click handlers only for books not already attached
    document.querySelectorAll('.book-item').forEach(item => {
        const bookId = parseInt(item.dataset.id);
        if (!attachedIds.has(bookId)) {
            item.style.cursor = 'pointer';
            item.addEventListener('click', () => attachExistingBook(bookId));
        } else {
            item.style.opacity = '0.6';
            item.style.cursor = 'default';
        }
    });
}

async function attachExistingBook(bookId) {
    const courseId = document.getElementById('course-id').value;
    try {
        const response = await fetch(`/courseBook/${courseId}/books/${bookId}/attach`, { method: 'POST' });
        if (!response.ok) throw new Error();
        showToast('کتاب با موفقیت به دوره اضافه شد', true);
        // Close modal and refresh book list
        closeExistingBooksModal();
        loadBooks(courseId);
    } catch (err) {
        showToast('خطا در افزودن کتاب', false);
    }
}

function closeExistingBooksModal() {
    document.getElementById('existing-books-modal').style.display = 'none';
}

// Add event listener for the search input
function initExistingBooks() {
    const searchInput = document.getElementById('search-books');
    if (searchInput) {
        let debounceTimer;
        searchInput.addEventListener('input', (e) => {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(() => {
                loadAllBooks(e.target.value);
            }, 300);
        });
    }
    document.getElementById('add-existing-book-btn')?.addEventListener('click', openExistingBooksModal);
    document.querySelector('.existing-books-close')?.addEventListener('click', closeExistingBooksModal);
    document.getElementById('cancel-existing-books')?.addEventListener('click', closeExistingBooksModal);
    window.addEventListener('click', (e) => {
        const modal = document.getElementById('existing-books-modal');
        if (e.target === modal) closeExistingBooksModal();
    });
}

function closeBookModal() {
    document.getElementById('book-modal').style.display = 'none';
    currentBookId = 0;
}


function initBooks() {
    initExistingBooks()
    document.getElementById('delete-from-course-btn').addEventListener('click', detachBookFromCourse);
    document.getElementById('delete-permanently-btn').addEventListener('click', deleteBookPermanently);
    document.getElementById('cancel-delete-book-btn').addEventListener('click', closeBookDeleteModal);
    document.querySelector('.delete-book-close').addEventListener('click', closeBookDeleteModal);
    window.addEventListener('click', (e) => {
        const modal = document.getElementById('delete-book-modal');
        if (e.target === modal) closeBookDeleteModal();
    });
    
    // Open modal for new book
    const addBtn = document.getElementById('add-book-btn');
    if (addBtn) {
        addBtn.addEventListener('click', () => openBookModal(0));
    } else {
        console.error('add-book-btn not found');
    }
    

    // book-form submit handler
    document.getElementById('book-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const courseId = document.getElementById('course-id').value;
        const bookId = document.getElementById('book-id').value;
        const title = document.getElementById('book-title').value;
        const downloadUrl = document.getElementById('book-download-url').value;
        const bookFile = document.getElementById('book-file').files[0];
        const thumbnailFile = document.getElementById('book-thumbnail').files[0];

        // Validate title
        if (!title.trim()) {
            showToast('لطفاً عنوان کتاب را وارد کنید', false);
            return;
        }

        if (!bookId || bookId === '0') {  // New book
            if (!thumbnailFile) {
                showToast('لطفاً تصویر جلد کتاب را انتخاب کنید', false);
                return;
            }
            if (!bookFile && !downloadUrl) {
                showToast('لطفاً فایل PDF کتاب را آپلود کنید یا لینک دانلود وارد کنید', false);
                return;
            }
        }

        const formData = new FormData();
        formData.append('title', title);
        if (downloadUrl) formData.append('download_url', downloadUrl);
        if (bookFile) formData.append('book_file', bookFile);
        if (thumbnailFile) formData.append('thumbnail', thumbnailFile);

        let url, method;
        if (bookId && bookId !== '0') {
            url = `/books/${bookId}`;
            method = 'PUT';
            formData.append('id', bookId);
        } else {
            url = `/courses/${courseId}/books`;
            method = 'POST';
        }

        // Show progress containers
        const thumbProgressContainer = document.getElementById('thumb-progress-container');
        const pdfProgressContainer = document.getElementById('pdf-progress-container');
        if (thumbnailFile) thumbProgressContainer.style.display = 'block';
        if (bookFile) pdfProgressContainer.style.display = 'block';

        // Reset progress bars
        const thumbBar = document.getElementById('thumb-progress-bar');
        const pdfBar = document.getElementById('pdf-progress-bar');
        if (thumbBar) thumbBar.style.width = '0%';
        if (pdfBar) pdfBar.style.width = '0%';

        // Use XMLHttpRequest to track upload progress
        const xhr = new XMLHttpRequest();
        xhr.open(method, url, true);

        // Track progress for file upload
        xhr.upload.addEventListener('progress', (event) => {
            if (event.lengthComputable) {
                const percentComplete = (event.loaded / event.total) * 100;
                // We have two files – we can't distinguish which one is being uploaded.
                // Simpler: update both bars? But that's not accurate.
                // Better: we'll show a single overall progress or update both together.
                // For simplicity, update both with the same percent (the browser interleaves both files).
                if (thumbProgressContainer.style.display === 'block') {
                    thumbBar.style.width = percentComplete + '%';
                    thumbBar.textContent = Math.round(percentComplete) + '%';
                }
                if (pdfProgressContainer.style.display === 'block') {
                    pdfBar.style.width = percentComplete + '%';
                    pdfBar.textContent = Math.round(percentComplete) + '%';
                }
            }
        });

        xhr.onload = () => {
            if (xhr.status >= 200 && xhr.status < 300) {
                showToast(bookId ? 'کتاب ویرایش شد' : 'کتاب اضافه شد', true);
                closeBookModal();
                if (courseId) loadBooks(courseId);
            } else {
                showToast('خطا در ذخیره کتاب', false);
            }
            // Hide progress bars
            thumbProgressContainer.style.display = 'none';
            pdfProgressContainer.style.display = 'none';
        };

        xhr.onerror = () => {
            showToast('خطا در ذخیره کتاب (مشکل شبکه)', false);
            thumbProgressContainer.style.display = 'none';
            pdfProgressContainer.style.display = 'none';
        };

        xhr.send(formData);
});

    
    document.querySelector('.book-close').addEventListener('click', closeBookModal);
    document.getElementById('book-cancel').addEventListener('click', closeBookModal);
    window.addEventListener('click', (e) => {
        if (e.target === document.getElementById('book-modal')) closeBookModal();
    });

    const bookModal = document.getElementById('book-modal');
    if (bookModal) {
        document.querySelector('.book-close')?.addEventListener('click', closeBookModal);
        document.getElementById('book-cancel')?.addEventListener('click', closeBookModal);
        window.addEventListener('click', (e) => {
            if (e.target === bookModal) closeBookModal();
        });
    }
}