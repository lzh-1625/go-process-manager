// users Data Page
export default [
  {
    path: "/process",
    component: () => import("@/views/process/Process.vue"),
    meta: {
      requiresAuth: true,
      layout: "landing",
      category: "Data",
    },
  },
];
