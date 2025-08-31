<script setup lang="ts">
import { login } from "@/api/login";
import router from "~/src/router";

const isLoading = ref(false);
const isSignInDisabled = ref(false);

const refLoginForm = ref();
const username = ref("");
const password = ref("");
const isFormValid = ref(true);

// show password field
const showPassword = ref(false);
const handleLogin = async () => {
  const { valid } = await refLoginForm.value.validate();
  if (valid) {
    isLoading.value = true;
    isSignInDisabled.value = true;
    login({
      username: username.value,
      password: password.value,
    }).then((e) => {
      isLoading.value = false;
      isSignInDisabled.value = false;

      if (e.code === 0) {
        localStorage.setItem("token", e.data?.token!);
        localStorage.setItem("role", e.data?.role.toString()!);
        localStorage.setItem("name", e.data?.username!);
        router.push("/");
      }
    });
  }
};

// error provider
const errorProvider = ref(false);
const errorProviderMessages = ref("");

const error = ref(false);
const errorMessages = ref("");
const resetErrors = () => {
  error.value = false;
  errorMessages.value = "";
};
</script>
<template>
  <v-card color="white" class="pa-3 ma-3" elevation="3">
    <v-card-title class="my-4 text-h4">
      <span class="flex-fill"> Xcon </span>
    </v-card-title>
    <v-card-subtitle>请使用你的游戏账号登入</v-card-subtitle>
    <!-- sign in form -->

    <v-card-text>
      <v-form
        ref="refLoginForm"
        class="text-left"
        v-model="isFormValid"
        lazy-validation
      >
        <v-text-field
          ref="refAccount"
          v-model="username"
          required
          :error="error"
          :label="$t('login.account')"
          density="default"
          variant="underlined"
          color="primary"
          bg-color="#fff"
          name="username"
          outlined
          validateOn="blur"
          placeholder=""
          @keyup.enter="handleLogin"
          @change="resetErrors"
        ></v-text-field>
        <v-text-field
          ref="refPassword"
          v-model="password"
          :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
          :type="showPassword ? 'text' : 'password'"
          :error="error"
          :error-messages="errorMessages"
          :label="$t('login.password')"
          placeholder=""
          density="default"
          variant="underlined"
          color="primary"
          bg-color="#fff"
          name="password"
          outlined
          validateOn="blur"
          @change="resetErrors"
          @keyup.enter="handleLogin"
          @click:append-inner="showPassword = !showPassword"
        ></v-text-field>
        <v-btn
          :loading="isLoading"
          :disabled="isSignInDisabled"
          block
          size="x-large"
          color="primary"
          @click="handleLogin"
          class="mt-2"
          >{{ $t("login.button") }}</v-btn
        >

        <div v-if="errorProvider" class="error--text my-2">
          {{ errorProviderMessages }}
        </div>
      </v-form></v-card-text
    >
  </v-card>
</template>
