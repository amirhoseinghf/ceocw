document.addEventListener('DOMContentLoaded', function() {
    const navItems = Array.from(document.querySelectorAll('.nav-item'));
    const panes = Array.from(document.querySelectorAll('.content-pane'));

    function activateSection(item, preserveScroll = true) {
        if (!item) return;
        const scrollY = window.scrollY;
        const sectionId = item.id.replace('-nav', '-content');

        navItems.forEach(navItem => navItem.classList.toggle('active', navItem === item));
        panes.forEach(pane => {
            pane.classList.toggle('active', pane.id === sectionId);
        });

        if (preserveScroll) {
            requestAnimationFrame(() => window.scrollTo({ top: scrollY, left: 0, behavior: 'auto' }));
        }
    }

    navItems.forEach(item => {
        item.addEventListener('click', function() {
            activateSection(this);
        });
    });

    activateSection(document.querySelector('.nav-item.active') || navItems[0], false);
});
