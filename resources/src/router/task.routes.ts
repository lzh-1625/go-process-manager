// users Data Page
export default [
  {
    path: "/task",
    component: () => import("@/views/task/Task.vue"),
    meta: {
      requiresAuth: true,
      layout: "landing",
      category: "Data",
    },
  },
];
