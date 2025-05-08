import React, { useEffect } from 'react';
import { SubmitHandler, useForm } from 'react-hook-form';
import isEmail from 'validator/lib/isEmail';
import { Tooltip } from 'react-tooltip';

interface AdminFormProps {
  onSubmit: (email: string) => void;
  message: { type: 'success' | 'error'; message: string | null };
  clearMessage: () => void;
}

const AdminForm: React.FC<AdminFormProps> = ({ onSubmit, message, clearMessage }) => {
  interface Inputs {
    email: string;
  }

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
    reset,
  } = useForm<Inputs>();

  const onSubmitHandler: SubmitHandler<Inputs> = async (data) => {
    const email = data.email.toLowerCase();

    if (!isEmail(email)) {
      setError('email', { type: 'error', message: 'Please provide a valid email address.' });
      return;
    }

    try {
      // sends email to adminPage
      await onSubmit(email);
      reset();
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    if (message.message) {
      const timer = setTimeout(() => clearMessage(), 4000);
      return () => clearTimeout(timer);
    }
  }, [message.message, clearMessage]);

  return (
    <>
      <h1 className="admin-title">Admin Page</h1>
      <div className="admin-form-container">
        <form onSubmit={handleSubmit(onSubmitHandler)}>
          <header className="infoText">Give a user admin rights:</header>
          <div className="input-button-wrapper ">
            <input
              id="email"
              {...register('email', {
                required: 'Email is required',
                pattern: {
                  value: /^\S+@\S+$/i,
                  message: 'Invalid email format',
                },
              })}
              placeholder="email"
              aria-label="Enter email to create a admin user"
              className={errors.email ? 'error-border' : ''}
            />

            <button type="submit" aria-label="Submit admin user" data-tooltip-id="submit-tooltip" data-tooltip-content={'Submit admin user'}>
              <i className="pi pi-save" />
            </button>
            <button
              type="button"
              onClick={() => {
                reset();
              }}
              data-tooltip-id="clear-tooltip"
              data-tooltip-content="Clear inputfield"
            >
              <i className="pi pi-eraser"></i>
            </button>
          </div>
          {errors.email && <p className="info-panel warning">{errors.email.message}</p>}
          {message.message && <p className={message.type === 'success' ? 'success' : 'warning'}>{message.message}</p>}
        </form>
        <div>
          <Tooltip id="submit-tooltip" />
          <Tooltip id="clear-tooltip" />
        </div>
      </div>
    </>
  );
};

export default AdminForm;
