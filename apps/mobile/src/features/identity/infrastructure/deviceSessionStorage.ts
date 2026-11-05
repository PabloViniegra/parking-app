import AsyncStorage from '@react-native-async-storage/async-storage';
import { getRandomBytes } from 'expo-crypto';
import * as SecureStore from 'expo-secure-store';

import { createEncryptedStorage, type KeyValueStore } from './encryptedStorage';

const keyStore: KeyValueStore = {
  getItem: (key) => SecureStore.getItemAsync(key),
  setItem: (key, value) =>
    SecureStore.setItemAsync(key, value, { keychainAccessible: SecureStore.AFTER_FIRST_UNLOCK }),
  removeItem: (key) => SecureStore.deleteItemAsync(key),
};

export const deviceSessionStorage = createEncryptedStorage({
  keyStore,
  dataStore: AsyncStorage,
  randomBytes: getRandomBytes,
});
