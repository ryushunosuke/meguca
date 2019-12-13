use web_sys;

// Get JS window global
pub fn window() -> web_sys::Window {
	web_sys::window().expect("window undefined")
}

// Get page document
pub fn document() -> web_sys::Document {
	window().document().expect("document undefined")
}

// Get page body
pub fn body() -> web_sys::HtmlElement {
	document().body().expect("body undefined")
}