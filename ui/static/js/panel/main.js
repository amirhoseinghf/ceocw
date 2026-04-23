// main.js
document.addEventListener('DOMContentLoaded', () => {
    // Initialize all modules
    initTeachers();
    initSemesters();
    initBooks();
    initCourses();

    // Preload dropdowns for courses
    loadSemesterOptions();
    loadTeacherOptions();

    // Tab switching
    const navItems = document.querySelectorAll('.nav-item');
    const tabs = {
        dashboard: document.getElementById('dashboard-content'),
        teachers: document.getElementById('teachers-content'),
        semesters: document.getElementById('semesters-content'),
        courses: document.getElementById('courses-content')
    };

    function switchTab(tabId) {
        Object.keys(tabs).forEach(id => {
            const tabContent = tabs[id];
            const navItem = document.querySelector(`.nav-item[data-tab="${id}"]`);
            if (tabContent) tabContent.classList.remove('active');
            if (navItem) navItem.classList.remove('active');
        });
        const activeTab = tabs[tabId];
        const activeNav = document.querySelector(`.nav-item[data-tab="${tabId}"]`);
        if (activeTab) activeTab.classList.add('active');
        if (activeNav) activeNav.classList.add('active');

        if (tabId === 'teachers') loadTeachers();
        if (tabId === 'semesters') loadSemesters();
        if (tabId === 'courses') loadCourses();
    }

    navItems.forEach(item => {
        item.addEventListener('click', (e) => {
            e.preventDefault();
            const tab = item.getAttribute('data-tab');
            if (tab) switchTab(tab);
        });
    });

    // Initially load default tab if needed (teachers or dashboard)
    // You can call switchTab('dashboard') or let the HTML decide
});