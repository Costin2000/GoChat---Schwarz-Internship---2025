<template>
  <div>
    <nav class="navbar navbar-expand navbar-light bg-light px-3">
      <a class="navbar-brand" href="/">GoChat</a>
      <div class="navbar-nav">
        <router-link class="nav-link" to="/">Home</router-link>
        <router-link class="nav-link" to="/register">Register</router-link>
        <router-link class="nav-link" to="/login">Login</router-link>
      </div>
    </nav>

    <div class="container py-4">
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts">
</script>