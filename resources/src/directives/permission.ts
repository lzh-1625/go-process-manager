export default {
  mounted(el, binding) {
    // 获取用户角色，角色数字越小权限越大: 0=root, 1=admin, 2=user
    const userRole = parseInt(localStorage.getItem("role") ?? "2");
    // 指令绑定值是所需的最低权限级别
    const requiredRole = binding.value;

    // 如果用户角色数字大于要求的角色，说明权限不足，隐藏元素
    if (userRole > requiredRole) {
      el.parentNode && el.parentNode.removeChild(el);
    }
  },
};
