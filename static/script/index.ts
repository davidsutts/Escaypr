document.addEventListener("DOMContentLoaded", () => {
	var button = document.getElementById("logout-btn")
	if (button) {
		button.addEventListener("click", logout)
	}
})

function logout () {
	let url = "/logout"
	let xhr = new XMLHttpRequest;
	xhr.open("POST", url);

	xhr.onload = () => {
		if (xhr.status == 200) {
			window.location.replace("/login")
		}
	}

	xhr.send()
}