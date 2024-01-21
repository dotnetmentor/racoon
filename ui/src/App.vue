<template>
  <header class="container-fluid">
    <nav>
      <ul>
        <li><h1>Racoon UI</h1></li>
      </ul>
      <ul>
        <li><router-link to="/">Start</router-link></li>
        <li><router-link to="/configs">Configurations</router-link></li>
        <li><router-link to="/compare">Compare</router-link></li>
      </ul>
    </nav>
  </header>

  <main class="container-fluid">
    <router-view />
  </main>
</template>
