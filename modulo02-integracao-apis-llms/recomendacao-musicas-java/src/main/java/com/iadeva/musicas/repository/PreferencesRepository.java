package com.iadeva.musicas.repository;

import com.iadeva.musicas.model.UserPreferences;
import org.springframework.data.jpa.repository.JpaRepository;
import java.util.Optional;

public interface PreferencesRepository extends JpaRepository<UserPreferences, String> {
    Optional<UserPreferences> findByUserId(String userId);
}
