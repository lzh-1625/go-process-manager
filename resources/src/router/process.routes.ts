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
  {
    path: "/process/share-manage",
    component: () => import("@/views/share/ShareManage.vue"),
    meta: {
      requiresAuth: true,
      layout: "landing",
      category: "Data",
    },
  },
];
