// home.js
document.addEventListener('DOMContentLoaded', function() {
    const searchInput = document.getElementById('search-input');
    const suggestionsContainer = document.getElementById('suggestions');
    const latestContainer = document.getElementById('latest-courses');
    let allCoursesCache = null;
    let isLoadingSearch = false;

    async function fetchLatestCourses() {
        latestContainer.innerHTML = '<div class="no-results">در حال بارگذاری...</div>';
        try {
            const response = await fetch('/latest/courses');
            if (!response.ok) throw new Error();
            const courses = await response.json();
            renderLatest(courses);
        } catch (err) {
            latestContainer.innerHTML = '<div class="no-results">خطا در بارگذاری دوره‌ها</div>';
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
                         onerror="this.onerror=null;this.src='/static/img/course_placeholder.jpg';">
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

    document.addEventListener('click', function(e) {
        if (!searchInput.contains(e.target) && !suggestionsContainer.contains(e.target)) {
            suggestionsContainer.style.display = 'none';
        }
    });

    searchInput.addEventListener('input', handleSearchInput);
    searchInput.addEventListener('focus', function() {
        if (searchInput.value.length > 0) handleSearchInput();
    });

    fetchLatestCourses();
});