export default {
  mounted(el, binding) {
    if (binding.value < (localStorage.getItem("permission") ?? 0)) {
      el.parentNode && el.parentNode.removeChild(el);
    }
  },
};
