// home.js
document.addEventListener('DOMContentLoaded', function() {
    const searchInput = document.getElementById('search-input');
    const searchIcon = document.getElementById('search-icon');
    const suggestionsContainer = document.getElementById('suggestions');
    const latestContainer = document.getElementById('latest-courses');
    const recordingsContainer = document.getElementById('recordings-list');
    const recordingsSearch = document.getElementById('recordings-search');
    const recordingsSort = document.getElementById('recordings-sort');
    const recordingsCount = document.getElementById('recordings-count');
    const recordingHintModal = document.getElementById('recording-hint-modal');
    const recordingHintCourse = document.getElementById('recording-hint-course');
    let allCoursesCache = null;
    let allRecordings = [];
    let isLoadingSearch = false;

    // Make the search icon clickable
    if (searchIcon && searchInput) {
        searchIcon.addEventListener('click', function() {
            const query = searchInput.value.trim();
            if (query) {
                window.location.href = `/search?q=${encodeURIComponent(query)}`;
            } else {
                searchInput.focus();
            }
        });
        // Ensure cursor changes to pointer (CSS already does, but double-check)
        searchIcon.style.cursor = 'pointer';
    }

    // Load latest courses (only 4)
    async function fetchLatestCourses() {
        latestContainer.innerHTML = '<div class="no-results">در حال بارگذاری...</div>';
        try {
            const response = await fetch('/latest/courses');
            if (!response.ok) throw new Error();
            const courses = await response.json();
            renderLatest(courses);
        } catch (err) {
            latestContainer.innerHTML = '<div class="no-results">خطا در بارگذاری دوره‌ها</div>';
            console.error(err);
        }
    }

    function renderLatest(courses) {
        if (!courses.length) {
            latestContainer.innerHTML = '<div class="no-results">هیچ دوره‌ای یافت نشد.</div>';
            return;
        }
        const html = courses.map(course => `
            <div class="course-card">
                <a href="/course/${course.Slug}" class="card-link">
                    <img src="${course.ImageUrl ? escapeHtml(course.ImageUrl) : '/static/img/course-placeholder.jpg'}"
                         class="card-image"
                         alt="${escapeHtml(course.Title)}"
                         onerror="this.onerror=null;this.src='/static/img/course-placeholder.jpg';">
                    <div class="card-content">
                        <div class="card-title">${escapeHtml(course.Title)}</div>
                        <div class="card-teacher">${escapeHtml(course.TeacherName)}</div>
                        <div class="card-semester">${escapeHtml(course.SemesterName)}</div>
                    </div>
                </a>
            </div>
        `).join('');
        latestContainer.innerHTML = html;
    }

    async function ensureAllCoursesLoaded() {
        if (allCoursesCache) return allCoursesCache;
        if (isLoadingSearch) return null;
        isLoadingSearch = true;
        try {
            const response = await fetch('/courses');
            if (!response.ok) throw new Error();
            allCoursesCache = await response.json();
            return allCoursesCache;
        } catch (err) {
            console.error('Failed to load courses for search', err);
            return [];
        } finally {
            isLoadingSearch = false;
        }
    }

    function filterSuggestions(courses, searchTerm) {
        const term = searchTerm.trim().toLowerCase();
        if (!term) return [];
        return courses.filter(course => 
            course.Title.toLowerCase().includes(term) ||
            (course.TeacherName && course.TeacherName.toLowerCase().includes(term)) ||
            (course.SemesterName && course.SemesterName.toLowerCase().includes(term))
        ).slice(0, 8);
    }

    function renderSuggestions(suggestions) {
        if (suggestions.length === 0) {
            suggestionsContainer.style.display = 'none';
            return;
        }
        suggestionsContainer.style.display = 'block';
        const html = suggestions.map(s => `
            <div class="suggestion-item" data-slug="${s.Slug}">
                <div class="suggestion-title">${escapeHtml(s.Title)}</div>
                <div class="suggestion-meta">${escapeHtml(s.TeacherName)} | ${escapeHtml(s.SemesterName)}</div>
            </div>
        `).join('');
        suggestionsContainer.innerHTML = html;

        document.querySelectorAll('.suggestion-item').forEach(el => {
            el.addEventListener('click', () => {
                const slug = el.dataset.slug;
                window.location.href = `/course/${slug}`;
            });
        });
    }

    async function handleSearchInput() {
        const term = searchInput.value;
        if (term.length === 0) {
            suggestionsContainer.style.display = 'none';
            return;
        }
        const courses = await ensureAllCoursesLoaded();
        if (courses) {
            const suggestions = filterSuggestions(courses, term);
            renderSuggestions(suggestions);
        }
    }

    // Search on Enter key
    searchInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            e.preventDefault();
            const query = searchInput.value.trim();
            if (query) {
                window.location.href = `/search?q=${encodeURIComponent(query)}`;
            }
        }
    });

    searchInput.addEventListener('input', handleSearchInput);
    searchInput.addEventListener('focus', function() {
        if (searchInput.value.length > 0) handleSearchInput();
    });

    // Close suggestions when clicking outside
    document.addEventListener('click', function(e) {
        if (!searchInput.contains(e.target) && !suggestionsContainer.contains(e.target)) {
            suggestionsContainer.style.display = 'none';
        }
    });

    function performSearchFromIcon() {
    const searchInput = document.getElementById('search-input');
    if (searchInput) {
        const query = searchInput.value.trim();
        if (query) {
            window.location.href = `/search?q=${encodeURIComponent(query)}`;
        }
    }
}

    async function fetchLatestTeachers() {
    const container = document.getElementById('latest-teachers');
    container.innerHTML = '<div class="no-results">در حال بارگذاری...</div>';
    try {
        const response = await fetch('/latest/teachers');
        if (!response.ok) throw new Error();
        const teachers = await response.json();
        renderTeachers(teachers);
    } catch (err) {
        container.innerHTML = '<div class="no-results">خطا در بارگذاری مدرسین</div>';
        console.error(err);
    }
}

    function renderTeachers(teachers) {
        const container = document.getElementById('latest-teachers');
        if (!teachers.length) {
            container.innerHTML = '<div class="no-results">هیچ مدرسی یافت نشد.</div>';
            return;
        }
        const html = teachers.map(teacher => `
            <div class="teacher-card">
                <a href="/teacher/${teacher.FirstNameEnglish.toLowerCase()}-${teacher.LastNameEnglish.toLowerCase().replace(/ /g, '-')}" class="teacher-card-link">
                    <img src="${teacher.ImageURL || '/static/img/teacher-placeholder.jpg'}" class="teacher-avatar" alt="${teacher.FirstName} ${teacher.LastName}" onerror="this.src='/static/img/teacher-placeholder.jpg';">
                    <div class="teacher-name">دکتر ${teacher.FirstName} ${teacher.LastName}</div>
                    <div class="teacher-slug">مدرس</div>
                </a>
            </div>
        `).join('');
        container.innerHTML = html;
    }

    function showRecordingHint(courseName) {
        if (!recordingHintModal || !recordingHintCourse || !courseName) return;
        recordingHintCourse.textContent = courseName;
        recordingHintModal.hidden = false;
        document.body.style.overflow = 'hidden';
    }

    function closeRecordingHint() {
        if (!recordingHintModal) return;
        recordingHintModal.hidden = true;
        document.body.style.overflow = '';
    }

    function arToPersian(value) {
        return String(value || '').replace(/[\u0643\u064a\u0649]/g, char => {
            if (char === '\u0643') return '\u06a9';
            return '\u06cc';
        });
    }

    function toPersianDigits(value) {
        return String(value).replace(/\d/g, digit => '۰۱۲۳۴۵۶۷۸۹'[digit]);
    }

    function normalizeSession(session) {
        const value = arToPersian(session).trim();
        if (!value || value === '—' || value === 'تست') return '—';
        const englishDigits = value.replace(/[۰-۹٠-٩]/g, digit =>
            String(digit.codePointAt(0) - (digit >= '٠' && digit <= '٩' ? 0x660 : 0x6f0))
        );
        if (/^\d+$/.test(englishDigits)) return `جلسه ${toPersianDigits(englishDigits)}`;
        return value;
    }

    function sessionRank(session) {
        const value = arToPersian(session).trim();
        if (!value || value === '—' || value === 'تست') return Number.POSITIVE_INFINITY;
        const normalized = value.replace(/[۰-۹٠-٩]/g, digit =>
            String(digit.codePointAt(0) - (digit >= '٠' && digit <= '٩' ? 0x660 : 0x6f0))
        );
        const numericMatch = normalized.match(/\d+/);
        if (numericMatch) return Number(numericMatch[0]);

        const wordRanks = {
            'صفر': 0,
            'اول': 1,
            'یک': 1,
            'دوم': 2,
            'دو': 2,
            'سوم': 3,
            'سه': 3,
            'چهارم': 4,
            'چهار': 4,
            'پنجم': 5,
            'پنج': 5,
            'ششم': 6,
            'شش': 6,
            'هفتم': 7,
            'هفت': 7,
            'هشتم': 8,
            'هشت': 8,
            'نهم': 9,
            'نه': 9,
            'دهم': 10,
            'ده': 10
        };
        const found = Object.entries(wordRanks).find(([word]) => value.includes(word));
        return found ? found[1] : Number.POSITIVE_INFINITY;
    }

    function formatFileSize(bytes) {
        const size = Number(bytes) || 0;
        if (!size) return '—';
        if (size >= 1073741824) return `${toPersianDigits((size / 1073741824).toFixed(1))} GB`;
        if (size >= 1048576) return `${toPersianDigits((size / 1048576).toFixed(1))} MB`;
        return `${toPersianDigits(Math.round(size / 1024))} KB`;
    }

    function formatRecordingDate(value) {
        if (!value) return '—';
        const date = new Date(value);
        if (Number.isNaN(date.getTime())) return '—';
        return date.toLocaleDateString('fa-IR');
    }

    function recordingSearchText(recording) {
        return [
            recording.course_name,
            recording.session,
            recording.name
        ].map(item => arToPersian(item).toLowerCase()).join(' ');
    }

    function groupRecordingsByCourse(recordings) {
        const groups = new Map();
        recordings.forEach(recording => {
            const courseName = recording.course_name || 'بدون عنوان';
            if (!groups.has(courseName)) {
                groups.set(courseName, {
                    courseName,
                    files: [],
                    newestTime: 0,
                    totalSize: 0
                });
            }
            const group = groups.get(courseName);
            group.files.push(recording);
            group.newestTime = Math.max(group.newestTime, new Date(recording.modified_at || 0).getTime() || 0);
            group.totalSize += Number(recording.size) || 0;
        });

        return Array.from(groups.values()).map(group => ({
            ...group,
            files: group.files.sort((a, b) =>
                sessionRank(a.session) - sessionRank(b.session) ||
                normalizeSession(a.session).localeCompare(normalizeSession(b.session), 'fa', { numeric: true }) ||
                new Date(a.modified_at || 0) - new Date(b.modified_at || 0)
            )
        }));
    }

    async function fetchRecordings() {
        if (!recordingsContainer) return;
        recordingsContainer.innerHTML = '<div class="no-results">در حال بارگذاری...</div>';
        try {
            const response = await fetch('/recordings');
            if (!response.ok) throw new Error();
            const recordings = await response.json();
            allRecordings = (recordings || []).map(recording => ({
                ...recording,
                course_name: arToPersian(recording.course_name),
                session: arToPersian(recording.session)
            }));
            renderRecordings();
        } catch (err) {
            recordingsContainer.innerHTML = '<div class="no-results">خطا در بارگذاری کلاس‌های ضبط شده</div>';
            if (recordingsCount) recordingsCount.textContent = '';
            console.error(err);
        }
    }

    function renderRecordings() {
        if (!recordingsContainer) return;
        const query = arToPersian(recordingsSearch?.value || '').trim().toLowerCase();
        const sort = recordingsSort?.value || 'date';
        const filtered = allRecordings.filter(recording => {
            const session = normalizeSession(recording.session);
            if (session === '—' && !recording.course_name && !recording.name) return false;
            if (!query) return true;
            return recordingSearchText(recording).includes(query);
        });

        const groups = groupRecordingsByCourse(filtered);
        groups.sort((a, b) => {
            switch (sort) {
                case 'count':
                    return b.files.length - a.files.length ||
                        a.courseName.localeCompare(b.courseName, 'fa', { numeric: true });
                case 'date':
                    return b.newestTime - a.newestTime ||
                        a.courseName.localeCompare(b.courseName, 'fa', { numeric: true });
                case 'course':
                default:
                    return a.courseName.localeCompare(b.courseName, 'fa', { numeric: true });
            }
        });

        if (recordingsCount) {
            recordingsCount.textContent = filtered.length
                ? `${toPersianDigits(filtered.length)} فایل`
                : '';
        }

        if (!filtered.length) {
            recordingsContainer.innerHTML = '<div class="no-results">هیچ فایلی یافت نشد.</div>';
            return;
        }

        recordingsContainer.innerHTML = groups.map(group => {
            const courseName = escapeHtml(group.courseName);
            const sessionCount = toPersianDigits(group.files.length);
            const totalSize = escapeHtml(formatFileSize(group.totalSize));
            const rows = group.files.map(recording => {
            const downloadUrl = escapeHtml(recording.download_url || '#');
            const session = escapeHtml(normalizeSession(recording.session));
            const fileName = escapeHtml(recording.name || '');
            const size = escapeHtml(formatFileSize(recording.size));
            const date = escapeHtml(formatRecordingDate(recording.modified_at));

            return `
                    <div class="recording-session-row">
                        <div class="recording-session-title">${session}</div>
                        <div class="recording-file-name">${fileName}</div>
                        <div class="recording-session-meta">
                        <span>${size}</span>
                        <span>${date}</span>
                    </div>
                        <div class="recording-session-actions">
                        <a href="${downloadUrl}" class="recording-download" target="_blank" rel="noopener">دانلود</a>
                        <button type="button" class="recording-copy" data-url="${downloadUrl}" title="کپی لینک">⛓</button>
                    </div>
                </div>
            `;
            }).join('');

            return `
                <section class="recording-course-group">
                    <div class="recording-course-header">
                        <div class="recording-course-name">${courseName}</div>
                        <div class="recording-course-summary">${sessionCount} جلسه · ${totalSize}</div>
                    </div>
                    <div class="recording-sessions-list">
                        ${rows}
                    </div>
                </section>
            `;
        }).join('');

        recordingsContainer.querySelectorAll('.recording-copy').forEach(button => {
            button.addEventListener('click', () => copyRecordingLink(button.dataset.url, button));
        });
    }

    function copyRecordingLink(url, button) {
        const done = () => {
            const previous = button.textContent;
            button.textContent = '✓';
            setTimeout(() => { button.textContent = previous; }, 1200);
        };
        if (navigator.clipboard?.writeText) {
            navigator.clipboard.writeText(url).then(done).catch(() => fallbackCopyRecording(url, done));
        } else {
            fallbackCopyRecording(url, done);
        }
    }

    function fallbackCopyRecording(url, done) {
        const textArea = document.createElement('textarea');
        textArea.value = url;
        textArea.style.cssText = 'position:fixed;opacity:0;top:0;left:0';
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();
        try {
            document.execCommand('copy');
            done();
        } catch (err) {
            window.prompt('لینک دانلود:', url);
        }
        document.body.removeChild(textArea);
    }

    if (recordingsSearch) recordingsSearch.addEventListener('input', renderRecordings);
    if (recordingsSort) recordingsSort.addEventListener('change', renderRecordings);
    document.querySelectorAll('[data-recording-hint-close]').forEach(element => {
        element.addEventListener('click', closeRecordingHint);
    });
    document.addEventListener('keydown', event => {
        if (event.key === 'Escape') closeRecordingHint();
    });
    const recordingCourseParam = new URLSearchParams(window.location.search).get('recording_course');
    if (recordingCourseParam) {
        showRecordingHint(recordingCourseParam);
    }

    fetchLatestCourses();
    fetchLatestTeachers();
    fetchRecordings();
});
