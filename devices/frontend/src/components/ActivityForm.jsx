import React, { useState } from 'react';
import axios from 'axios';

const ActivityForm = () => {
  const [location, setLocation] = useState('');
  const [password, setPassword] = useState('');
  const [screenshot, setScreenshot] = useState(null);
  const [message, setMessage] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    const formData = new FormData();
    formData.append('location', location);
    formData.append('password', password);
    formData.append('screenshot', screenshot);

    try {
      const response = await axios.post('/upload/', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });

      if (response.data.status === 'success') {
        setMessage('Activity added successfully');
      } else {
        setMessage('Failed to add activity');
      }
    } catch (err) {
      console.error('Error:', err);
      setMessage('Failed to add activity');
    }
  };

  return (
    <div>
      <h2>User Activity Submission</h2>
      <form onSubmit={handleSubmit}>
        <input
          type="text"
          placeholder="Location"
          value={location}
          onChange={(e) => setLocation(e.target.value)}
          required
        />
        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
        />
        <input
          type="file"
          accept="image/*"
          onChange={(e) => setScreenshot(e.target.files[0])}
          required
        />
        <button type="submit">Add Activity</button>
      </form>
      <p>{message}</p>
    </div>
  );
};

export default ActivityForm;
