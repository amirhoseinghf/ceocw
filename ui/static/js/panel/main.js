// main.js
document.addEventListener('DOMContentLoaded', () => {
    applyRoleVisibility();
    normalizeJalaliInputs(document);
    initFileDropzones(document);

    // Initialize all modules
    initTeachers();
    initSemesters();
    initBooks();
    initCourses();
    initNotes();
    initExams();
    initTAs();
    if (typeof initUsers === 'function') initUsers();
    if (typeof initCourseUsers === 'function') initCourseUsers();
    if (typeof initAnnouncements === 'function') initAnnouncements();

    initCourseDescription();
    initSlides();
    initAssignments();

    // Preload dropdowns for courses
    loadSemesterOptions();
    loadTeacherOptions();

    const tabs = {
        dashboard: document.getElementById('dashboard-content'),
        teachers:  document.getElementById('teachers-content'),
        semesters: document.getElementById('semesters-content'),
        courses:   document.getElementById('courses-content'),
        users:     document.getElementById('users-content')
    };

    // ── helpers ───────────────────────────────────────────────────────────
    function tabPath(tabId) {
        return tabId === 'dashboard' ? '/panel' : '/panel/' + tabId;
    }

    function parsePath(pathname) {
        // /panel               → { tab: 'dashboard', courseId: null }
        // /panel/courses       → { tab: 'courses',   courseId: null }
        // /panel/courses/5     → { tab: 'courses',   courseId: 5    }
        // /panel/semesters     → { tab: 'semesters', courseId: null }
        const stripped = pathname.replace(/^\/panel\/?/, ''); // "courses/5" or ""
        const parts = stripped.split('/').filter(Boolean);
        const tab = parts[0] || 'dashboard';
        const courseId = parts[1] ? parseInt(parts[1], 10) : null;
        return { tab, courseId };
    }

    // ── internal tab switch: no history push ──────────────────────────────
    function activateTab(tabId) {
        Object.keys(tabs).forEach(id => {
            const tabContent = tabs[id];
            const navItem = document.querySelector(`.nav-item[data-tab="${id}"]`);
            if (tabContent) tabContent.classList.remove('active');
            if (navItem)    navItem.classList.remove('active');
        });
        const activeTab = tabs[tabId];
        const activeNav = document.querySelector(`.nav-item[data-tab="${tabId}"]`);
        if (activeTab) activeTab.classList.add('active');
        if (activeNav) activeNav.classList.add('active');

        if (tabId === 'teachers') loadTeachers();
        if (tabId === 'semesters') loadSemesters();
        if (tabId === 'users' && typeof loadUsers === 'function') loadUsers();

        // Clicking "courses" tab always navigates to the list view
        if (tabId === 'courses') {
            const listDiv = document.getElementById('courses-list');
            const editDiv = document.getElementById('course-edit-panel');
            if (listDiv) listDiv.classList.remove('hidden');
            if (editDiv) editDiv.classList.add('hidden');
            loadCourses();
        }

        applyRoleVisibility();
    }

    // ── public API: pushes to browser history ─────────────────────────────
    window.panelSwitchTab = function(tabId) {
        activateTab(tabId);
        history.pushState({ tab: tabId }, '', tabPath(tabId));
    };

    window.panelOpenCourse = function(courseId) {
        // activateTab shows list first; showCourseManage immediately shows edit panel
        activateTab('courses');
        showCourseManage(courseId);
        history.pushState({ tab: 'courses', courseId }, '', '/panel/courses/' + courseId);
    };

    window.panelBackToCourses = function() {
        const listDiv = document.getElementById('courses-list');
        const editDiv = document.getElementById('course-edit-panel');
        if (listDiv) listDiv.classList.remove('hidden');
        if (editDiv) editDiv.classList.add('hidden');
        loadCourses();
        history.pushState({ tab: 'courses' }, '', '/panel/courses');
    };

    // ── sidebar nav clicks ────────────────────────────────────────────────
    document.querySelectorAll('.nav-item').forEach(item => {
        item.addEventListener('click', (e) => {
            e.preventDefault();
            const tab = item.getAttribute('data-tab');
            if (tab) window.panelSwitchTab(tab);
        });
    });

    // ── browser back / forward ────────────────────────────────────────────
    window.addEventListener('popstate', () => {
        const { tab, courseId } = parsePath(location.pathname);
        const safeTab = (tab && tabs[tab]) ? tab : 'dashboard';
        activateTab(safeTab);
        if (safeTab === 'courses' && courseId && !isNaN(courseId)) {
            showCourseManage(courseId);
        }
    });

    // ── initial navigation from URL path ─────────────────────────────────
    const { tab: initTab, courseId: initCourseId } = parsePath(location.pathname);
    const startTab = (initTab && tabs[initTab]) ? initTab : 'dashboard';
    activateTab(startTab);
    // Replace current history entry so back-button works correctly
    history.replaceState({ tab: startTab, courseId: initCourseId }, '', location.pathname);

    if (startTab === 'courses' && initCourseId && !isNaN(initCourseId)) {
        showCourseManage(initCourseId);
    }
});
