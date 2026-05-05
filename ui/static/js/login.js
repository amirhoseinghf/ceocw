// login.js
document.addEventListener('DOMContentLoaded', function() {
    const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.get('registered') === 'true') {
            showToast('ثبت‌نام با موفقیت انجام شد. لطفاً وارد شوید.', true);
            // Clean the URL without reloading
            history.replaceState({}, '', '/user/login');
        } else if (urlParams.get('denied') == 'true') {
            showToast('دسترسی به این بخش برای شما مقدور نمی‌باشد', false)
            history.replaceState({}, '', '/user/login');
        }
    const form = document.querySelector('.auth-form');
    if (!form) return;

    form.addEventListener('submit', function(e) {
        e.preventDefault();

        const email = document.getElementById('email').value.trim();
        const password = document.getElementById('password').value;

        if (!email) {
            showToast('لطفاً ایمیل خود را وارد کنید', false);
            return;
        }
        if (!isValidEmail(email)) {
            showToast('لطفاً یک ایمیل معتبر وارد کنید', false);
            return;
        }
        if (!password) {
            showToast('لطفاً رمز عبور را وارد کنید', false);
            return;
        }

        form.submit();
    });

    function isValidEmail(email) {
        const re = /^[^\s@]+@([^\s@.,]+\.)+[^\s@.,]{2,}$/;
        return re.test(email);
    }
});