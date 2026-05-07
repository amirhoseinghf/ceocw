document.addEventListener('DOMContentLoaded', function() {
    var form = document.getElementById('profile-form');
    var imageInput = document.getElementById('profile-image-path');
    var imageFileInput = document.getElementById('profile-image-file');
    var previewWrap = document.getElementById('profile-image-preview-wrap');
    var avatarPreview = document.getElementById('profile-avatar-preview');
    if (!form || !imageInput || !imageFileInput || !avatarPreview || !previewWrap) return;

    var updatePreview = function(value) {
        avatarPreview.src = value || '/static/img/teacher-placeholder.jpg';
        previewWrap.style.display = value ? 'block' : 'none';
    };

    if (typeof bindFileDropzone === 'function') bindFileDropzone(imageFileInput);
    imageFileInput.addEventListener('change', function() {
        var file = imageFileInput.files && imageFileInput.files[0];
        if (!file) {
            updatePreview(imageInput.value.trim());
            return;
        }
        var reader = new FileReader();
        reader.onload = function(e) {
            updatePreview(e.target.result);
        };
        reader.readAsDataURL(file);
    });
    avatarPreview.addEventListener('error', function() {
        avatarPreview.src = '/static/img/teacher-placeholder.jpg';
    });

    form.addEventListener('submit', async function(e) {
        e.preventDefault();
        var payload = {
            firstName: document.getElementById('profile-first-name').value.trim(),
            lastName: document.getElementById('profile-last-name').value.trim(),
            email: document.getElementById('profile-email').value.trim(),
            imagePath: imageInput.value.trim(),
            password: document.getElementById('profile-password').value
        };

        if (!payload.firstName || !payload.lastName || !payload.email) {
            showToast('نام، نام خانوادگی و ایمیل الزامی است', false);
            return;
        }

        try {
            var formData = new FormData();
            formData.append('first_name', payload.firstName);
            formData.append('last_name', payload.lastName);
            formData.append('email', payload.email);
            formData.append('password', payload.password);
            formData.append('image_path', payload.imagePath);
            var imageFile = imageFileInput.files && imageFileInput.files[0];
            if (imageFile) formData.append('user_image', imageFile);
            var response = await fetch('/user/profile', {
                method: 'PUT',
                body: formData
            });
            if (!response.ok) {
                var text = await response.text();
                throw new Error(text || 'خطا در ذخیره پروفایل');
            }
            var updated = await response.json();
            imageInput.value = updated.imagePath || '';
            document.getElementById('profile-password').value = '';
            imageFileInput.value = '';
            updatePreview(imageInput.value.trim());
            if (typeof refreshFileDropzones === 'function') refreshFileDropzones(form);
            showToast('پروفایل به‌روزرسانی شد', true);
        } catch (err) {
            showToast(err.message || 'خطا در ذخیره پروفایل', false);
        }
    });
});