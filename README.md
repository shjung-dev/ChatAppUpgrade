# Upgraded Chat

This project is an improved version of 1 to 1 Chat Application that includes message persistence, real-time communication, and group chat functionality. Unlike a previous simple 1-to-1 chat app, this app ensures that all messages remain stored and accessible even after refreshing the page or logging back in.

---

## Key Features

### 1. JWT Authentication
Users are securely authenticated using **access and refresh tokens** with claims, ensuring safe API requests and WebSocket connections. This allows the system to identify users sending messages in real-time.

### 2. Real-Time Friend Requests
Users can send friend requests to each other:  
- If both users are online, the request appears immediately on the receiver’s browser.  
- If the receiver is offline, the request is automatically shown when they next connect, ensuring seamless communication across sessions.

### 3. Message Persistence
All messages, both 1-to-1 and group chats, are stored in MongoDB. Chat history is preserved even after refreshing the page or logging back in, allowing users to continue conversations without interruption.

### 4. Group Chat Functionality
Users can create group chats with at least two other friends. Messages sent in group chats are also persistent and updated in real-time, enabling smooth collaborative communication.

> **Note:** Features like deleting messages, leaving groups, or removing friends were intentionally omitted. The project focuses on understanding real-time message persistence, differentiating between 1-to-1 and group chats, and handling real-time events such as friend requests.

---

## Architecture & Flow

- **Frontend:** Built with Next.js  
- **Backend:** Golang using Gin framework for HTTP requests  
- **Real-time Communication:** Gorilla WebSocket  
- **Authentication:** Golang-JWT for access & refresh tokens  
- **Database:** MongoDB  

**Flow Overview:**  
1. Users log in or sign up → credentials validated against the **User Table** → redirected to homepage.  
2. On homepage, the client automatically connects to the **WebSocket server** using JWT.  
3. Searching a user triggers a protected GET request, returning user info if found.  
4. Sending a friend request goes through WebSocket → updates the **Request Table** → notifies receiver in real-time.  
5. Accepting/rejecting requests updates the **Request Table** and, if accepted, the **Friend Table**.  
6. Sending messages: WebSocket updates **Conversation** and **Message Tables**, creating new conversations if needed.  
7. On homepage refresh or reconnect, existing conversations and incoming friend requests are fetched and displayed.

> **Deployment Note:** The frontend is hosted on Vercel, and the backend on Render.com. Render’s free plan may cause initial requests to delay up to ~50 seconds if the instance is inactive.

---

## Tech Stack

- **Frontend:** Next.js  
- **Backend:** Golang, Gin Framework  
- **Real-Time Messaging:** Gorilla WebSocket  
- **Authentication:** Golang-JWT (Access & Refresh Tokens)  
- **Database:** MongoDB  
- **Deployment:** Frontend → Vercel, Backend → Render.com  

---


