document.addEventListener("DOMContentLoaded", function () {
    var button = document.getElementById("logout-btn");
    if (button) {
        button.addEventListener("click", logout);
    }
});
function logout() {
    var url = "/logout";
    var xhr = new XMLHttpRequest;
    xhr.open("POST", url);
    xhr.onload = function () {
        if (xhr.status == 200) {
            window.location.replace("/login");
        }
    };
    xhr.send();
}
