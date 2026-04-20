document.addEventListener('DOMContentLoaded', function() {
    // Handle navigation clicks
    const navItems = document.querySelectorAll('.nav-item');
    navItems.forEach(item => {
        item.addEventListener('click', function() {
            // Remove active class from all items
            navItems.forEach(navItem => navItem.classList.remove('active'));
            
            // Add active class to clicked item
            this.classList.add('active');
            
            // Hide all content sections
            const announcementsContent = document.getElementById('announcements-content');
            const assignmentsContent = document.getElementById('assignments-content');
            const slidesContent = document.getElementById('slides-content');
            const notesContent = document.getElementById('notes-content');
            const examsContent = document.getElementById('exams-content');

            if (announcementsContent) {
                announcementsContent.style.display = 'none';
            }
            if (assignmentsContent) {
                assignmentsContent.style.display = 'none';
            }
            if (slidesContent) {
                slidesContent.style.display = 'none';
            }
            if (notesContent) {
                notesContent.style.display = 'none';
            }
            if (examsContent) {
                examsContent.style.display = 'none';
            }
            
            // Show the selected section
            const sectionId = this.id.replace('-nav', '-content');
            const contentElement = document.getElementById(sectionId);
            if (contentElement) {
                contentElement.style.display = 'block';
            }
        });
    });
    
    // Initialize with first section visible
    if (document.getElementById('announcements-nav') && document.getElementById('announcements-content')) {
        document.getElementById('announcements-content').style.display = 'block';
        document.getElementById('assignments-content').style.display = 'none';
        document.getElementById('slides-content').style.display = 'none';
        document.getElementById('notes-content').style.display = 'none';
         document.getElementById('exams-content').style.display = 'none';
    } else if (document.getElementById('assignments-nav') && document.getElementById('assignments-content')) {
        document.getElementById('announcements-content').style.display = 'none';
    } else if (document.getElementById('slides-nav') && document.getElementById('slides-content')) {
        document.getElementById('announcements-content').style.display = 'none';
    } else if (document.getElementById('notes-nav') && document.getElementById('notes-content')) {
        document.getElementById('announcements-content').style.display = 'none';
    } else if (document.getElementById('exams-nav') && document.getElementById('exams-content')) {
        document.getElementById('announcements-content').style.display = 'none';
    } 

    const header = document.querySelector(".nav-dropdown-header");
    const dropdown = document.getElementById("nav-dropdown");
    const selected = document.getElementById("dropdown-selected");

    if (!header || !dropdown) return;

    const activeItem = document.querySelector(".nav-item.active");
    if (activeItem) selected.textContent = activeItem.textContent.trim();

    // Toggle dropdown on header click
    header.addEventListener("click", () => {
        dropdown.classList.toggle("open");
    });

    // NEW: Close dropdown when any item is clicked
    document.querySelectorAll(".nav-item").forEach(item => {
        item.addEventListener("click", () => {
            dropdown.classList.remove("open");
            selected.textContent = item.textContent.trim(); // optional
        });
    });

});