// users Data Page
export default [
  {
    path: "/user",
    component: () => import("@/views/user/User.vue"),
    meta: {
      requiresAuth: true,
      layout: "landing",
      category: "Data",
    },
  },
];
