<template>
  <div class="auth-card">
    <h1 class="title">Create a New Account</h1>
    <p class="subtitle">Enter your details below to get started.</p>
    
    <form @submit.prevent="handleRegister" class="auth-form">
      <div class="input-group">
        <label for="firstName">First Name</label>
        <input id="firstName" type="text" v-model="form.firstName" placeholder="e.g., John" required>
      </div>
      <div class="input-group">
        <label for="lastName">Last Name</label>
        <input id="lastName" type="text" v-model="form.lastName" placeholder="e.g., Doe" required>
      </div>
      <div class="input-group">
        <label for="userName">Username</label>
        <input id="userName" type="text" v-model="form.userName" placeholder="e.g., john.doe" required>
      </div>
      <div class="input-group">
        <label for="email">Email Address</label>
        <input id="email" type="email" v-model="form.email" placeholder="e.g., email@example.com" required>
      </div>
      <div class="input-group">
        <label for="password">Password</label>
        <input id="password" type="password" v-model="form.password" placeholder="Minimum 6 characters" required>
      </div>

      <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>

      <button type="submit" class="auth-button" :disabled="isLoading">
        {{ isLoading ? 'Registering...' : 'Create Account' }}
      </button>
    </form>

    <p class="footer-text">
      Already have an account? 
      <RouterLink to="/login" class="link">Log in here</RouterLink>
    </p>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { useRouter, RouterLink } from 'vue-router';

// import { UserServiceClient } from '@/proto/services/user-base/proto/userbase_grpc_web_pb';
// import { User, CreateUserRequest } from '@/proto/services/user-base/proto/userbase_pb';

const router = useRouter();

const form = reactive({
  firstName: '',
  lastName: '',
  userName: '',
  email: '',
  password: '',
});

const isLoading = ref(false);
const errorMessage = ref('');

async function handleRegister() {
  errorMessage.value = '';

  // Frontend validation
  if (form.password.length < 6) {
    errorMessage.value = 'Password must be at least 6 characters long.';
    return;
  }
  
  isLoading.value = true;

  try {
    console.log('Submitting registration data:', form);
    
    // --- gRPC-Web API call example (uncomment after running `make proto-frontend`) ---
    /*
    const client = new UserServiceClient('http://localhost:8080', null, null);
    const user = new User();
    user.setFirstName(form.firstName);
    user.setLastName(form.lastName);
    user.setUserName(form.userName);
    user.setEmail(form.email);
    user.setPassword(form.password);
    const request = new CreateUserRequest();
    request.setUser(user);
    const response = await client.createUser(request, {});
    console.log('User created successfully:', response.toObject());
    router.push('/chat');
    */

    // Simulate a successful API call for now
    await new Promise(resolve => setTimeout(resolve, 1000));
    alert('Account created successfully! (Simulation) Redirecting...');
    router.push('/chat');

  } catch (error) {
    console.error('Registration failed:', error);
    errorMessage.value = 'An unexpected error occurred. Please try again.';
  } finally {
    isLoading.value = false;
  }
}
</script>

<style scoped>
.auth-card {
  background-color: white;
  padding: 40px;
  border-radius: 12px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
  width: 100%;
  max-width: 420px;
  text-align: center;
}
.title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 8px;
  color: #1a202c;
}
.subtitle {
  font-size: 14px;
  color: #718096;
  margin-bottom: 32px;
}
.auth-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  text-align: left;
}
.input-group {
  display: flex;
  flex-direction: column;
}
.input-group label {
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 6px;
  color: #4a5568;
}
.input-group input {
  padding: 10px 12px;
  border: 1px solid #cbd5e0;
  border-radius: 8px;
  font-size: 14px;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.input-group input:focus {
  outline: none;
  border-color: #4299e1;
  box-shadow: 0 0 0 2px rgba(66, 153, 225, 0.5);
}
.auth-button {
  padding: 12px;
  border: none;
  border-radius: 8px;
  background-color: #4299e1;
  color: white;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
  margin-top: 8px;
}
.auth-button:hover {
  background-color: #3182ce;
}
.auth-button:disabled {
  background-color: #a0aec0;
  cursor: not-allowed;
}
.error-message {
  color: #e53e3e;
  font-size: 13px;
  text-align: center;
  margin-top: -8px;
  margin-bottom: 8px;
}
.footer-text {
  margin-top: 24px;
  font-size: 14px;
  color: #718096;
}
.link {
  color: #4299e1;
  font-weight: 500;
  text-decoration: none;
}
.link:hover {
  text-decoration: underline;
}
</style>

